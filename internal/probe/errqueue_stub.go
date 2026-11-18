//go:build !linux

package probe

import (
	"context"

	"github.com/cloudapp3/vmlg/internal/enrich"
)

func runPlatformFallbackMTR(_ context.Context, _ MTRRequest, _ *enrich.Enricher, _ func(MTRSnapshot)) (MTRSnapshot, bool, error) {
	return MTRSnapshot{}, false, nil
}
