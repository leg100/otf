document.addEventListener('alpine:init', () => {
  Alpine.data('dropdown', (existing = [], available = []) => ({
    open: false,
    close(focusAfter) {
      if (! this.open) return
      this.open = false
      focusAfter && focusAfter.focus()
    },
    search: '',
    existing: existing,
    available: available,
    get filterAvailable() {
      return this.available?.filter(
        i => i.name.includes(this.search)
      ).slice(0, 3)
    },
    get showPanel() {
      return (this.open && this.filterAvailable?.length > 0)
    },
    addItem(item) {
      // move item from available to existing
      this.available = this.available.filter(
        i => i !== item
      )
      this.existing.push(item)
      // hide dropdown box
      this.close()
    },
    deleteItem(item) {
      // move item from existing to available
      this.existing = this.existing.filter(
        i => i !== item
      )
      this.available.push(item)
    },
  }))
})
