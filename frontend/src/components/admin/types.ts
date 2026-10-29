export interface Column {
  key: string
  label: string
  align?: 'left' | 'right' | 'center'
  /** Shown without its label as the card title on mobile. */
  primary?: boolean
  width?: string
}
