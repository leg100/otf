document.addEventListener('alpine:init', () => {
  Alpine.data('agent_pool_workspaces', (allowed = [], available = []) => ({
    open: false,
    close(focusAfter) {
      if (! this.open) return
      this.open = false
      focusAfter && focusAfter.focus()
    },
    search: '',
    allowed: allowed,
    available: available,
    get filterAvailable() {
      return this.available?.filter(
        i => i.name.includes(this.search)
      ).slice(0, 3)
    },
    get showPanel() {
      return (this.open && this.filterAvailable?.length > 0)
    },
    addWorkspace(workspace) {
      // move workspace from available to allowed
      this.available = this.available.filter(
        i => i !== workspace
      )
      this.allowed.push(workspace)
      // hide dropdown box
      this.close()
    },
    deleteWorkspace(workspace) {
      // move workspace from allowed to available
      this.allowed = this.allowed.filter(
        i => i !== workspace
      )
      this.available.push(workspace)
    },
  }))
})
