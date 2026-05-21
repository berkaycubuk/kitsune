package report

import (
	"encoding/json"
	"io"

	"github.com/berkaycubuk/kitsune/internal/runner"
)

func WriteJSON(w io.Writer, r *runner.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}
