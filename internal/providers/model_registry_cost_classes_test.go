package providers

import "testing"

func TestModelRegistryListByCostClasses(t *testing.T) {
	r := NewStaticModelMetadataRegistry(
		ModelMetadata{Provider: "x", Model: "free", CostClass: CostClassFree},
		ModelMetadata{Provider: "x", Model: "low", CostClass: CostClassLow},
		ModelMetadata{Provider: "x", Model: "high", CostClass: CostClassHigh},
	)

	models := r.ListByCostClasses("x", CostClassFree, CostClassLow)
	if len(models) != 2 {
		t.Fatalf("models = %d, want 2", len(models))
	}
}
