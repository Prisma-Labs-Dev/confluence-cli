package confluence

import "testing"

func TestBuildPageSearchCQL(t *testing.T) {
	t.Run("query with space key", func(t *testing.T) {
		cql, err := buildPageSearchCQL(PageSearchOptions{Query: "slotting", SpaceKey: "SC"})
		if err != nil {
			t.Fatalf("buildPageSearchCQL: %v", err)
		}
		want := `type=page AND text ~ "slotting" AND space="SC"`
		if cql != want {
			t.Fatalf("cql = %q, want %q", cql, want)
		}
	})

	t.Run("raw cql", func(t *testing.T) {
		cql, err := buildPageSearchCQL(PageSearchOptions{CQL: `space = "TNLTA" AND title ~ "OutSystems"`})
		if err != nil {
			t.Fatalf("buildPageSearchCQL: %v", err)
		}
		want := `space = "TNLTA" AND title ~ "OutSystems"`
		if cql != want {
			t.Fatalf("cql = %q, want %q", cql, want)
		}
	})

	t.Run("query and cql conflict", func(t *testing.T) {
		_, err := buildPageSearchCQL(PageSearchOptions{Query: "slotting", CQL: `space = "SC"`})
		if err == nil {
			t.Fatal("expected conflict error")
		}
	})
}
