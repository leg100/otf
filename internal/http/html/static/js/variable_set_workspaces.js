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
        i => i.Name.includes(this.search)
      ).slice(0, 3)
    },
    get showPanel() {
      return (this.open && this.filterAvailable?.length > 0)
    },
    addWorkspace(workspace) {
      // move workspace from available to existing
      this.available = this.available.filter(
        i => i !== workspace
      )
      this.existing.push(workspace)
      // hide dropdown box
      this.close()
    },
    deleteWorkspace(workspace) {
      // move workspace from existing to available
      this.existing = this.existing.filter(
        i => i !== workspace
      )
      this.available.push(workspace)
    },
  }))
})
