package codemapbench

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

func splitHoldout(links []Link, config Config) ([]Link, []Link, string, error) {
	known, err := normalizeKnownLinks(links)
	if err != nil {
		return nil, nil, "", err
	}
	if len(known) == 0 {
		return nil, nil, "", errors.New("benchmark requires at least one known link")
	}

	seed, count, err := resolveHoldout(config, len(known))
	if err != nil {
		return nil, nil, "", err
	}

	ranked := append([]Link(nil), known...)
	sort.Slice(ranked, func(i, j int) bool {
		left := holdoutHash(seed, ranked[i])
		right := holdoutHash(seed, ranked[j])
		if left == right {
			return linkKey(ranked[i]) < linkKey(ranked[j])
		}
		return left < right
	})

	hiddenSet := make(map[string]struct{}, count)
	for _, link := range ranked[:count] {
		hiddenSet[linkKey(link)] = struct{}{}
	}

	visible := make([]Link, 0, len(known)-count)
	hidden := make([]Link, 0, count)
	for _, link := range known {
		if _, ok := hiddenSet[linkKey(link)]; ok {
			hidden = append(hidden, link)
			continue
		}
		visible = append(visible, link)
	}
	return visible, hidden, seed, nil
}

func resolveHoldout(config Config, total int) (string, int, error) {
	seed := strings.TrimSpace(config.Seed)
	if seed == "" {
		seed = DefaultSeed
	}
	if config.HoldoutCount < 0 {
		return "", 0, errors.New("holdout count cannot be negative")
	}
	if config.HoldoutCount > 0 && config.HoldoutFraction > 0 {
		return "", 0, errors.New("set holdout count or holdout fraction, not both")
	}

	count := config.HoldoutCount
	if count == 0 {
		fraction := config.HoldoutFraction
		if fraction == 0 {
			fraction = 0.2
		}
		if fraction <= 0 || fraction > 1 {
			return "", 0, errors.New("holdout fraction must be greater than zero and at most one")
		}
		count = int(math.Ceil(float64(total) * fraction))
	}
	if count > total {
		return "", 0, fmt.Errorf("holdout count %d exceeds %d known links", count, total)
	}
	return seed, count, nil
}

func holdoutHash(seed string, link Link) string {
	sum := sha256.Sum256([]byte(seed + "\x00" + linkKey(link)))
	return string(sum[:])
}
