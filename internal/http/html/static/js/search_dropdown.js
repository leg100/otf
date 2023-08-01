document.addEventListener('alpine:init', () => {
  Alpine.data('search_dropdown', (existing = [], available = []) => ({
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
        i => i.includes(this.search)
      ).slice(0, 3)
    },
    get isNew() {
      return this.search !== '' && !this.available?.includes(this.search) && !this.existing?.includes(this.search)
    },
    get showPanel() {
        if (this.open) {
          if (this.existing?.includes(this.search)) return true
          if (this.isNew) return true
          if (this.filterAvailable?.length > 0) return true
        }
        return false
    },
    get showAlreadyAdded() {
        return this.existing?.includes(this.search)
    },
  }))
})
