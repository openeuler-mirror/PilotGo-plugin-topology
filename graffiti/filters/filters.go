package filters

import (
	"regexp"

	"github.com/pmylund/go-cache"

	"github.com/skydive-project/skydive/graffiti/getter"
)

var regexpCache *cache.Cache

// Eval evaluates a filter
func (f *Filter) Eval(g getter.Getter) bool {
	if f.BoolFilter != nil {
		return f.BoolFilter.Eval(g)
	}
	if f.TermStringFilter != nil {
		return g.MatchString(f.TermStringFilter.Key, func(s string) bool { return s == f.TermStringFilter.Value })
	}
	if f.TermInt64Filter != nil {
		return g.MatchInt64(f.TermInt64Filter.Key, func(i int64) bool { return i == f.TermInt64Filter.Value })
	}
	if f.TermBoolFilter != nil {
		return g.MatchBool(f.TermBoolFilter.Key, func(b bool) bool { return b == f.TermBoolFilter.Value })
	}
	if f.GtInt64Filter != nil {
		return g.MatchInt64(f.GtInt64Filter.Key, func(i int64) bool { return i > f.GtInt64Filter.Value })
	}
	if f.LtInt64Filter != nil {
		return g.MatchInt64(f.LtInt64Filter.Key, func(i int64) bool { return i < f.LtInt64Filter.Value })
	}
	if f.GteInt64Filter != nil {
		return g.MatchInt64(f.GteInt64Filter.Key, func(i int64) bool { return i >= f.GteInt64Filter.Value })
	}
	if f.LteInt64Filter != nil {
		return g.MatchInt64(f.LteInt64Filter.Key, func(i int64) bool { return i <= f.LteInt64Filter.Value })
	}
	if f.RegexFilter != nil {
		return g.MatchString(f.RegexFilter.Key, func(s string) bool {
			re, found := regexpCache.Get(f.RegexFilter.Value)
			if !found {
				re = regexp.MustCompile(f.RegexFilter.Value)
				regexpCache.Set(f.RegexFilter.Value, re, cache.DefaultExpiration)
			}

			return re.(*regexp.Regexp).MatchString(s)
		})
	}
	if f.NullFilter != nil {
		if _, err := g.GetField(f.NullFilter.Key); err == nil {
			return false
		}
		return true
	}
	if f.IPV4RangeFilter != nil {
		return g.MatchString(f.IPV4RangeFilter.Key, func(s string) bool {
			re, found := regexpCache.Get(f.IPV4RangeFilter.Value)
			if !found {
				// ignore error at this point should have been check in the contructor
				regex, _ := IPV4CIDRToRegex(f.IPV4RangeFilter.Value)
				re = regexp.MustCompile(regex)
				regexpCache.Set(f.IPV4RangeFilter.Value, re, cache.DefaultExpiration)
			}

			return re.(*regexp.Regexp).MatchString(s)
		})
	}

	return true
}

// Eval evaluates a boolean (not, and, or) filter
func (b *BoolFilter) Eval(g getter.Getter) bool {
	for _, filter := range b.Filters {
		result := filter.Eval(g)
		if b.Op == BoolFilterOp_NOT && !result {
			return true
		}
		if b.Op == BoolFilterOp_AND && !result {
			return false
		} else if b.Op == BoolFilterOp_OR && result {
			return true
		}
	}
	return b.Op == BoolFilterOp_AND || len(b.Filters) == 0
}
