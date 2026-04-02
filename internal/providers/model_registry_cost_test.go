package providers

import "testing"

func TestListByCostClass(t *testing.T) {
	r := NewDefaultModelMetadataRegistry()
	models := r.ListByCostClass("openai", CostClassMedium)
	if len(models) == 0 {
		t.Fatalf("expected openai models with medium-or-lower cost class")
	}
	for _, m := range models {
		if costClassRank(m.CostClass) > costClassRank(CostClassMedium) {
			t.Fatalf("model %s exceeds ceiling", m.Model)
		}
	}
}
