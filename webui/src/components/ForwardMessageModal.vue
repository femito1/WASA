<!-- File: webui/src/components/ForwardMessageModal.vue -->
<template>
    <div class="modal-overlay">
      <div class="modal-content">
        <h5>Forward Message</h5>
        <div class="form-group mb-3">
          <label for="targetConv">Select Conversation:</label>
          <select id="targetConv" v-model="targetConversationId" class="form-select">
            <option v-for="conv in conversations" :key="conv.id" :value="conv.id">
              {{ conv.name || 'Unnamed Conversation' }}
            </option>
          </select>
        </div>
        <div class="d-flex justify-content-end gap-2">
          <button class="btn btn-secondary" @click="emit('close')">Cancel</button>
          <button class="btn btn-primary" @click="forward">Forward</button>
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import { ref, onMounted } from 'vue'
  import axios from '../services/axios'
  import jwtDecode from 'jwt-decode'
  import { useRouter } from 'vue-router'
  
  const props = defineProps({
    message: { type: Object, required: true }
  })
  
  // Define emits explicitly
  const emit = defineEmits(['forward', 'close'])
  
  const targetConversationId = ref(null)
  const conversations = ref([])
  
  const router = useRouter()
  // Get the current user ID from the token.
  const token = localStorage.getItem('authToken')
  if (!token) {
    router.push({ name: 'Login' })
    throw new Error('No token found')
  }
  const decoded = jwtDecode(token)
  const userId = decoded.user_id
  
  async function fetchConversations() {
    try {
      // GET /users/:id/conversations
      const response = await axios.get(`/users/${userId}/conversations`)
      conversations.value = response.data
      if (conversations.value.length) {
        targetConversationId.value = conversations.value[0].id
      }
    } catch (err) {
      console.error(err)
    }
  }
  
  function forward() {
    // Emit the selected conversation ID for forwarding.
    emit('forward', targetConversationId.value)
  }
  
  onMounted(() => {
    fetchConversations()
  })
  </script>
  
  <style scoped>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0,0,0,0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }
  .modal-content {
    background: white;
    padding: 16px;
    border-radius: 8px;
    min-width: 300px;
  }
  </style>
  