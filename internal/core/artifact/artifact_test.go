package artifact

import "testing"

func TestArtifactIsPresent(t *testing.T) {
	present := Artifact{Status: StatusPresent}
	missing := Artifact{Status: StatusMissing}
	if !present.IsPresent() {
		t.Error("present.IsPresent() = false")
	}
	if missing.IsPresent() {
		t.Error("missing.IsPresent() = true")
	}
}

func TestArtifactIsMissing(t *testing.T) {
	reqMissing := Artifact{Required: true, Status: StatusMissing}
	optMissing := Artifact{Required: false, Status: StatusMissing}
	if !reqMissing.IsMissing() {
		t.Error("required+missing.IsMissing() = false")
	}
	if optMissing.IsMissing() {
		t.Error("optional+missing.IsMissing() = true")
	}
}

func TestArtifactIsPlaceholder(t *testing.T) {
	ph := Artifact{Required: true, Path: "", Status: StatusMissing}
	real := Artifact{Required: true, Path: "files/cad/part.step", Status: StatusPresent}
	if !ph.IsPlaceholder() {
		t.Error("placeholder.IsPlaceholder() = false")
	}
	if real.IsPlaceholder() {
		t.Error("real.IsPlaceholder() = true")
	}
}

func TestArtifactKinds(t *testing.T) {
	kinds := []Kind{KindCAD, KindDrawing, KindManufacturing, KindImage, KindCertificate, KindEvidence, KindGenerated, KindDocument}
	for _, k := range kinds {
		if k == "" {
			t.Error("empty kind")
		}
	}
}

func TestArtifactStatuses(t *testing.T) {
	if StatusPresent == "" || StatusMissing == "" || StatusOutdated == "" {
		t.Error("status constants must be non-empty")
	}
}
