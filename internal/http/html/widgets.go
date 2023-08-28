package html

const (
	NarrowDropDown DropDownUIWidth = "narrow"
	WideDropDown   DropDownUIWidth = "wide"
)

type (
	// DropdownUI populates a search/dropdown UI component.
	DropdownUI struct {
		// Name to send along with value in the POST form
		Name string
		// Existing values to NOT show in the dropdown
		Existing []string
		// Available values to show in the dropdown
		Available []string
		// Action is the form action URL
		Action string
		// Placeholder to show in the input element.
		Placeholder string
		// Width: "narrow" or "wide"
		Width DropDownUIWidth
	}

	DropDownUIWidth string
)
