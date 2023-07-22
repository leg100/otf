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
      return this.available.filter(
        i => i.includes(this.search)
      ).slice(0, 3)
    },
    get isNew() {
      return this.search !== '' && !this.available.includes(this.search) && !this.existing.includes(this.search)
    },
  }))
})
