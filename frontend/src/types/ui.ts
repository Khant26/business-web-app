export interface SelectOption<V> {
  value: V
  label: string
  disabled?: boolean
}

export interface TabItem<V extends string = string> {
  value: V
  label: string
  count?: number
}
