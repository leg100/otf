document.addEventListener("alpine:init", () => {
	Alpine.data("combobox", (options = []) => ({
		options: options,
		isOpen: false,
		openedWithKeyboard: false,
		selectedOptions: [],
		setLabelText() {
			const count = this.selectedOptions.length;

			// if there are no selected options
			if (count === 0) return "Please Select";

			// if there is only one selected option
			return this.selectedOptions.join(", ");
		},
		highlightFirstMatchingOption(pressedKey) {
			// if Enter pressed, do nothing
			if (pressedKey === "Enter") return;

			// find and focus the option that starts with the pressed key
			const option = this.options.find((item) =>
				item.toLowerCase().startsWith(pressedKey.toLowerCase()),
			);
			if (option) {
				const index = this.options.indexOf(option);
				const allOptions = document.querySelectorAll(".combobox-option");
				if (allOptions[index]) {
					allOptions[index].focus();
				}
			}
		},
		handleOptionToggle(option) {
			if (option.checked) {
				this.selectedOptions.push(option);
			} else {
				// remove the unchecked option from the selectedOptions array
				this.selectedOptions = this.selectedOptions.filter(
					(opt) => opt !== option,
				);
			}
			// set the value of the hidden field to the selectedOptions array
			this.$refs.hiddenTextField.value = this.selectedOptions;
		},
	}));
});
