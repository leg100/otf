document.addEventListener('alpine:init', () => {
  Alpine.data('search_dropdown', (items = []) => ({
    open: false,
    close(focusAfter) {
      if (! this.open) return
      this.open = false
      focusAfter && focusAfter.focus()
    },
    search: '',
    items: items,
    get filterItems() {
      return this.items.filter(
        i => i.includes(this.search)
      ).slice(0, 3)
    },
    get exactMatch() {
      return this.search === '' || this.items.includes(this.search)
    },
  }))
})
