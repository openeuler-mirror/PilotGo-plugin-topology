import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export const useTopoStore = defineStore('topo', () => {
  const nodeData = ref({} as any)
  return { nodeData }
})
