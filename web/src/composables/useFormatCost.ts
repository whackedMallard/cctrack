export function formatCostDisplay(value: number): string {
  if (value < 0.01) return '$' + value.toFixed(4)
  return '$' + value.toFixed(2)
}

export function formatCostPrecise(value: number): string {
  return '$' + value.toFixed(4)
}

export function formatTokens(n: number): string {
  if (n < 1000) return String(n)
  if (n < 1_000_000) return (n / 1000).toFixed(1) + 'K'
  return (n / 1_000_000).toFixed(1) + 'M'
}

export function formatTokensRaw(n: number): string {
  return n.toLocaleString()
}

export function formatModel(model: string): string {
  return model
    .replace('claude-', '')
    .replace(/-\d{8}$/, '')
    .replace(/-/g, ' ')
}

export function formatDate(iso: string): string {
  if (!iso) return '—'
  const d = new Date(iso)
  const now = new Date()
  const isToday = d.toDateString() === now.toDateString()
  if (isToday) {
    return d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })
  }
  return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short' })
}
