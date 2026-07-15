export type User = {
  id: string
  email: string
  alias: string
  role: 'user' | 'admin'
  dateJoined: string
}
