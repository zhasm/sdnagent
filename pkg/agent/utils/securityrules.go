package utils

import (
	"fmt"
	"strings"

	"yunion.io/yunioncloud/pkg/util/secrules"
)

type SecurityRule struct {
	r          *secrules.SecurityRule
	ovsMatches []string
}

func NewSecurityRule(s string) (*SecurityRule, error) {
	s = strings.TrimSpace(s)
	r, err := secrules.ParseSecurityRule(s)
	if err != nil {
		return nil, err
	}
	return &SecurityRule{r: r}, nil
}

func (sr *SecurityRule) Direction() string {
	return sr.r.Direction
}

func (sr *SecurityRule) OvsMatches() []string {
	if sr.ovsMatches != nil {
		return sr.ovsMatches
	}
	var nwField string
	var tpField string
	var protoMatch string
	var nwMatch string
	var tpMatch []string

	r := sr.r
	switch r.Direction {
	case secrules.DIR_IN:
		nwField = "nw_src="
		tpField = "tp_src="
	case secrules.DIR_OUT:
		nwField = "nw_dst="
		tpField = "tp_dst="
	}
	if net := r.IPNet.String(); net != "0.0.0.0/0" {
		if ones, bits := r.IPNet.Mask.Size(); ones == 32 && bits == 32 {
			nwMatch = nwField + r.IPNet.IP.String()
		} else {
			nwMatch = nwField + net
		}
	}
	switch r.Protocol {
	case secrules.PROTO_ANY:
		protoMatch = "ip"
	case secrules.PROTO_TCP, secrules.PROTO_UDP:
		if len(r.Ports) > 0 {
			for _, p := range r.Ports {
				tpMatch = append(tpMatch, tpField+fmt.Sprintf("%d", p))
			}
		} else if r.PortStart > 0 && r.PortStart < r.PortEnd {
			ms := PortRangeToMasks(r.PortStart, r.PortEnd)
			for _, m := range ms {
				var vs string
				// NOTE both start and end should never be zero, the
				// check is here just in case
				if m[0] == 0 {
					vs = "0"
				} else {
					vs = fmt.Sprintf("0x%x", m[0])
				}
				if m[1] == 0 {
					break
				}
				tpMatch = append(tpMatch, fmt.Sprintf("%s%s/0x%x", tpField, vs, m[1]))
			}
		}
		fallthrough
	default:
		protoMatch = r.Protocol
	}

	var m string
	if len(nwMatch) > 0 {
		m = protoMatch + "," + nwMatch
	} else {
		m = protoMatch
	}
	if len(tpMatch) == 0 {
		sr.ovsMatches = []string{m}
	} else {
		ms := []string{}
		for _, tpm := range tpMatch {
			ms = append(ms, m+","+tpm)
		}
		sr.ovsMatches = ms
	}
	return sr.ovsMatches
}

func (sr *SecurityRule) OvsActionAllow() bool {
	return sr.r.Action == secrules.ACTION_ALLOW
}

func (sr *SecurityRule) IsAllowAny() bool {
	return sr.r.IsAllowAny()
}

// TODO squash neighbouring rules of the same direction
type SecurityRules struct {
	inRules       []*SecurityRule
	outRules      []*SecurityRule
	inOvsMatches  []string
	outOvsMatches []string
	inAllowAny    bool
	outAllowAny   bool
}

func (sr *SecurityRules) rulesString(srs []*SecurityRule) string {
	v := []string{}
	for _, r := range srs {
		v = append(v, r.r.String())
	}
	return strings.Join(v, "; ")
}

func (sr *SecurityRules) InRulesString() string {
	return sr.rulesString(sr.inRules)
}

func (sr *SecurityRules) OutRulesString() string {
	return sr.rulesString(sr.outRules)
}

func NewSecurityRules(s string) (*SecurityRules, error) {
	inRules := []*SecurityRule{}
	outRules := []*SecurityRule{}
	in, inAllowAny := false, false
	out, outAllowAny := false, false
	srs := strings.Split(s, ";")
	for _, sr := range srs {
		sr = strings.TrimSpace(sr)
		if len(sr) == 0 {
			continue
		}
		r, err := NewSecurityRule(sr)
		if err != nil {
			// TODO err wrapper
			return nil, err
		}
		switch r.Direction() {
		case secrules.DIR_IN:
			if !in {
				if r.IsAllowAny() {
					inAllowAny = true
				}
				in = true
			}
			inRules = append(inRules, r)
		case secrules.DIR_OUT:
			if !out {
				if r.IsAllowAny() {
					outAllowAny = true
				}
				out = true
			}
			outRules = append(outRules, r)
		}
	}
	// These can only happen when users have secgroup specified but rules
	// for each direction can be empty.
	//
	// In the case where no secgroup was assigned, default security_rules
	// "in:allow_any; out:allow_any" will be used by the caller
	if len(inRules) == 0 {
		r, _ := NewSecurityRule("in:deny any")
		inRules = append(inRules, r)
	}
	if len(outRules) == 0 {
		r, _ := NewSecurityRule("out:allow any")
		outRules = append(outRules, r)
		outAllowAny = true
	}
	return &SecurityRules{
		inRules:     inRules,
		outRules:    outRules,
		inAllowAny:  inAllowAny,
		outAllowAny: outAllowAny,
	}, nil
}

func (sr *SecurityRules) InAllowAny() bool {
	return sr.inAllowAny
}

func (sr *SecurityRules) OutAllowAny() bool {
	return sr.outAllowAny
}

func PortRangeToMasks(s, e uint16) [][2]uint16 {
	r := [][2]uint16{}
	if s == e {
		r = append(r, [2]uint16{s, ^uint16(0)})
		return r
	}
	sp, ep := uint32(s), uint32(e)
	ep = ep + 1
	for sp < ep {
		b := uint32(1)
		for (sp+b) <= ep && (sp&(b-1)) == 0 {
			b <<= 1
		}
		b >>= 1
		r = append(r, [2]uint16{uint16(sp), uint16(^(b - 1))})
		sp = sp + b
	}
	return r
}