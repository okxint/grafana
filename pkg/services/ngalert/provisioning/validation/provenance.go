package validation

import (
	"github.com/grafana/grafana/pkg/apimachinery/utils"
	"github.com/grafana/grafana/pkg/services/ngalert/models"
)

// CanUpdateManagerInRuleGroup checks if a manager can be updated for a rule group and its alerts.
// Preserves the same transition semantics as CanUpdateProvenanceInRuleGroup:
//   - same kind is always allowed
//   - unmanaged (unknown kind) stored → allow any incoming manager
//   - resetting to unmanaged → only allowed from api-like managers, not file/repo managers
func CanUpdateManagerInRuleGroup(stored, incoming utils.ManagerProperties) bool {
	if stored.Kind == incoming.Kind {
		return true
	}
	if stored.Kind == utils.ManagerKindUnknown {
		return true
	}
	if incoming.Kind == utils.ManagerKindUnknown {
		// Can reset to unmanaged only from api-style managers (not file/repo).
		switch stored.Kind { //nolint:staticcheck
		case utils.ManagerKindClassicAPI, utils.ManagerKindClassicConvertedPrometheus,
			utils.ManagerKindTerraform, utils.ManagerKindKubectl:
			return true
		}
		return false
	}
	return false
}

// CanUpdateProvenanceInRuleGroup checks if a provenance can be updated for a rule group and its alerts.
// ReplaceRuleGroup function intends to replace an entire rule group: inserting, updating, and removing rules.
//
// Deprecated: use CanUpdateManagerInRuleGroup for new code. This function remains for non-rule resources
// (receivers, templates, mute timings, routes) that have not yet migrated to ManagerProperties.
func CanUpdateProvenanceInRuleGroup(storedProvenance, provenance models.Provenance) bool {
	// Same provenance is always allowed
	if storedProvenance == provenance {
		return true
	}

	// Can always update stored ProvenanceNone
	if storedProvenance == models.ProvenanceNone {
		return true
	}

	// Can reset to ProvenanceNone from specific provenances
	if provenance == models.ProvenanceNone {
		return storedProvenance == models.ProvenanceAPI ||
			storedProvenance == models.ProvenanceConvertedPrometheus
	}

	return false
}

type ProvenanceStatusTransitionValidator = func(from, to models.Provenance) error

// ValidateProvenanceRelaxed checks if the transition of provenance status from `from` to `to` is allowed.
// Applies relaxed checks that prevents only transition from any status to `none`.
// Returns ErrProvenanceChangeNotAllowed if transition is not allowed
func ValidateProvenanceRelaxed(from, to models.Provenance) error {
	if from == models.ProvenanceNone { // allow any transition from none
		return nil
	}
	if to == models.ProvenanceNone { // allow any transition to none unless it's from "none" either
		return MakeErrProvenanceChangeNotAllowed(from, to)
	}
	return nil
}

// ValidateProvenanceOfDependentResources returns a list of allowed provenance statuses for dependent resources
// in the case when they need to be updated when the resource they depend on is changed.
func ValidateProvenanceOfDependentResources(parentProvenance models.Provenance) func(childProvenance models.Provenance) bool {
	return func(childProvenance models.Provenance) bool {
		return parentProvenance == childProvenance || childProvenance == models.ProvenanceNone
	}
}
