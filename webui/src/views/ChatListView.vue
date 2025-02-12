<!-- File: webui/src/views/ChatListView.vue -->
<template>
    <div class="container mt-3">
      <h1>Your Conversations</h1>
      <LoadingSpinner :loading="loading">
        <div v-if="!loading">
          <ul class="list-group">
            <ConversationItem
              v-for="conversation in conversations"
              :key="conversation.id"
              :conversation="conversation"
              @open="openConversation"
            />
          </ul>
        </div>
      </LoadingSpinner>
      <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
      <button class="btn btn-primary mt-3" @click="showNewConversationModal">New Conversation</button>

      <!-- New Conversation Modal -->
      <div v-if="showModal" class="modal-overlay">
        <div class="modal-content">
          <h5>Start New Conversation</h5>
          <div class="mb-3">
            <label class="form-label">Select Contact:</label>
            <select v-model="selectedContactId" class="form-select">
              <option value="">Choose a contact...</option>
              <option v-for="contact in contacts" :key="contact.id" :value="contact.id">
                {{ contact.username }}
              </option>
            </select>
          </div>
          <div class="d-flex justify-content-end gap-2">
            <button class="btn btn-secondary" @click="showModal = false">Cancel</button>
            <button class="btn btn-primary" @click="createConversation" :disabled="!selectedContactId">
              Start Conversation
            </button>
          </div>
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import { ref, onMounted } from 'vue'
  import { useRouter } from 'vue-router'
  import axios from '../services/axios'
  import LoadingSpinner from '../components/LoadingSpinner.vue'
  import ErrorMsg from '../components/ErrorMsg.vue'
  import ConversationItem from '../components/ConversationItem.vue'
  import jwtDecode from 'jwt-decode'
  
  const conversations = ref([])
  const contacts = ref([])
  const loading = ref(false)
  const errorMsg = ref(null)
  const showModal = ref(false)
  const selectedContactId = ref('')
  const router = useRouter()
  
  const token = localStorage.getItem('authToken')
  if (!token) {
    router.push({ name: 'Login' })
    throw new Error('No authentication token found.')
  }
  
  const decodedToken = jwtDecode(token)
  const userId = decodedToken.user_id
  
  async function fetchContacts() {
    try {
      const response = await axios.get(`/users/${userId}/contacts`)
      contacts.value = response.data
    } catch (err) {
      errorMsg.value = err.response?.data?.error || err.toString()
    }
  }
  
  async function fetchConversations() {
    loading.value = true
    errorMsg.value = null
    try {
      const response = await axios.get(`/users/${userId}/conversations`)
      conversations.value = response.data
    } catch (err) {
      errorMsg.value = err.response?.data?.error || err.toString()
    }
    loading.value = false
  }
  
  function showNewConversationModal() {
    if (contacts.value.length === 0) {
      errorMsg.value = "You need to add contacts first before starting a conversation"
      return
    }
    showModal.value = true
  }
  
  async function createConversation() {
    if (!selectedContactId.value) {
      errorMsg.value = "Please select a contact"
      return
    }

    try {
      const contact = contacts.value.find(c => c.id === selectedContactId.value)
      const conversationName = contact ? contact.username : 'New Conversation'
      
      const response = await axios.post(`/users/${userId}/conversations`, {
        name: conversationName,
        members: [selectedContactId.value]
      })
      
      showModal.value = false
      selectedContactId.value = ''
      await fetchConversations()
      
      router.push({ name: 'Chat', params: { convId: response.data.id } })
    } catch (err) {
      if (err.response?.status === 403) {
        errorMsg.value = "You can only start conversations with your contacts"
      } else if (err.response?.status === 400) {
        errorMsg.value = "Invalid request. Please check your input."
      } else {
        errorMsg.value = err.response?.data?.error || "Failed to create conversation"
      }
    }
  }
  
  function openConversation(conv) {
    router.push({ name: 'Chat', params: { convId: conv.id } })
  }
  
  onMounted(() => {
    fetchConversations()
    fetchContacts()
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
    width: 90%;
    max-width: 400px;
  }
  </style>
  