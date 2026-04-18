package probes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerVPC(r *provider.Registry, cli *shell.Client) {
	const typ = "vpc"
	r.Register(typ, map[string]provider.ProbeFn{
		"available": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"ec2", "describe-vpcs",
				"--vpc-ids", req.Name,
				"--query", "Vpcs[0].State",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			return provider.Result{
				Value:  strings.EqualFold(text, "available"),
				Raw:    text,
				Status: provider.StatusOk,
			}, nil
		},
		"ip_utilization": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			out, err := cli.Run(ctx,
				"ec2", "describe-subnets",
				"--filters", "Name=vpc-id,Values="+req.Name,
				"--query", "Subnets[].{cidr:CidrBlock,free:AvailableIpAddressCount}",
				"--output", "json")
			if err != nil {
				return provider.Result{}, err
			}
			var subnets []struct {
				Cidr string `json:"cidr"`
				Free int    `json:"free"`
			}
			if err := json.Unmarshal(out, &subnets); err != nil {
				return provider.Result{}, fmt.Errorf("%w: unable to parse subnets json: %v", provider.ErrProtocol, err)
			}
			if len(subnets) == 0 {
				return provider.FloatResult(0), nil
			}
			var total, used int
			for _, s := range subnets {
				capacity, err := cidrCapacity(s.Cidr)
				if err != nil {
					return provider.Result{}, fmt.Errorf("%w: bad CIDR %q: %v", provider.ErrProtocol, s.Cidr, err)
				}
				total += capacity
				used += capacity - s.Free
			}
			if total == 0 {
				return provider.FloatResult(0), nil
			}
			return provider.FloatResult(float64(used) / float64(total) * 100), nil
		},
	})
}

// cidrCapacity returns the number of usable host addresses in an IPv4
// CIDR as AWS counts them. AWS reserves 5 addresses per subnet (network,
// broadcast, VPC router, DNS, future). Matches AvailableIpAddressCount.
func cidrCapacity(cidr string) (int, error) {
	i := strings.LastIndexByte(cidr, '/')
	if i < 0 {
		return 0, fmt.Errorf("missing prefix length")
	}
	var prefix int
	if _, err := fmt.Sscanf(cidr[i+1:], "%d", &prefix); err != nil {
		return 0, err
	}
	if prefix < 0 || prefix > 32 {
		return 0, fmt.Errorf("invalid prefix /%d", prefix)
	}
	raw := 1 << (32 - prefix)
	if raw <= 5 {
		return 0, nil
	}
	return raw - 5, nil
}
