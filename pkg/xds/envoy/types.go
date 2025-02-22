package envoy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type ClusterSubset struct {
	ClusterName       string
	Weight            uint32
	Tags              Tags
	IsExternalService bool
}

type Tags map[string]string

func (t Tags) WithoutTag(tag string) Tags {
	result := Tags{}
	for tagName, tagValue := range t {
		if tag != tagName {
			result[tagName] = tagValue
		}
	}
	return result
}

func (t Tags) WithTags(keysAndValues ...string) Tags {
	result := Tags{}
	for tagName, tagValue := range t {
		result[tagName] = tagValue
	}
	for i := 0; i < len(keysAndValues); {
		key, value := keysAndValues[i], keysAndValues[i+1]
		result[key] = value
		i += 2
	}
	return result
}

func (t Tags) Keys() []string {
	var keys []string
	for key := range t {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (t Tags) String() string {
	var pairs []string
	for _, key := range t.Keys() {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, t[key]))
	}
	return strings.Join(pairs, ",")
}

func TagsFromString(tagsString string) (Tags, error) {
	result := Tags{}
	tagPairs := strings.Split(tagsString, ",")
	for _, pair := range tagPairs {
		split := strings.Split(pair, "=")
		if len(split) != 2 {
			return nil, errors.New("invalid format of tags, pairs should be separated by , and key should be separated from value by =")
		}
		result[split[0]] = split[1]
	}
	return result, nil
}

func DistinctTags(tags []Tags) []Tags {
	used := map[string]bool{}
	var result []Tags
	for _, tag := range tags {
		str := tag.String()
		if !used[str] {
			result = append(result, tag)
			used[str] = true
		}
	}
	return result
}

type Cluster struct {
	subsets            []ClusterSubset
	hasExternalService bool
}

func (c *Cluster) Add(subset ClusterSubset) {
	c.subsets = append(c.subsets, subset)
	if subset.IsExternalService {
		c.hasExternalService = true
	}
}

func (c *Cluster) Tags() []Tags {
	var result []Tags
	for _, info := range c.subsets {
		result = append(result, info.Tags)
	}
	return result
}

func (c *Cluster) HasExternalService() bool {
	return c.hasExternalService
}

func (c *Cluster) Subsets() []ClusterSubset {
	return c.subsets
}

type Clusters map[string]*Cluster

func (c Clusters) ClusterNames() []string {
	var keys []string
	for key := range c {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (c Clusters) Add(infos ...ClusterSubset) {
	for _, info := range infos {
		if c[info.ClusterName] == nil {
			c[info.ClusterName] = &Cluster{}
		}
		c[info.ClusterName].Add(info)
	}
}

func (c Clusters) Get(name string) *Cluster {
	return c[name]
}

func (c Clusters) Tags(name string) []Tags {
	return c[name].Tags()
}
