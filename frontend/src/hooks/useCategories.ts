import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'

export function useCategories() {
  return useQuery({
    queryKey: ['categories'],
    queryFn: api.categories.list,
    staleTime: 5 * 60 * 1000,
  })
}
