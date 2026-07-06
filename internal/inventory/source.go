package inventory

import (
	"context"

	"github.com/kyoh86/nvim-plugin-triage/internal/plugin"
)

type Source interface {
	List(context.Context) ([]plugin.Plugin, error)
}
