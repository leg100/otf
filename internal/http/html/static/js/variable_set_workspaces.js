document.addEventListener('alpine:init', () => {
  Alpine.data('variable_set_workspaces', (existing = [], available = []) => ({
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
  }))
})
