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
      <button class="btn btn-primary mt-3" @click="createConversation">New Conversation</button>
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
  
  // Reactive variables for conversation data, loading state, and errors.
  const conversations = ref([])
  const loading = ref(false)
  const errorMsg = ref(null)
  const router = useRouter()
  
  // Extract the JWT token from localStorage.
  const token = localStorage.getItem('authToken')
  if (!token) {
    // If no token is found, redirect the user to the Login view.
    router.push({ name: 'Login' })
    throw new Error('No authentication token found.')
  }
  
  // Decode the token to extract the user ID. Adjust the field name based on your JWT payload.
  const decodedToken = jwtDecode(token)
  const userId = decodedToken.user_id  // Ensure your backend sets the user ID under "user_id" in the token payload.
  
  // Fetch the user's conversations from the backend.
  async function fetchConversations() {
    loading.value = true
    errorMsg.value = null
    try {
      // API endpoint: GET /users/:id/conversations
      const response = await axios.get(`/users/${userId}/conversations`)
      conversations.value = response.data
    } catch (err) {
      errorMsg.value = err.response?.data?.error || err.toString()
    }
    loading.value = false
  }
  
  // Navigate to the Chat view when a conversation is selected.
  function openConversation(conv) {
    router.push({ name: 'Chat', params: { convId: conv.id } })
  }
  
  // Create a new conversation.
  async function createConversation() {
    const conversationName = prompt('Enter a conversation name:', 'New Conversation')
    if (!conversationName) return
    try {
      await axios.post(`/users/${userId}/conversations`, { name: conversationName, members: [] })
      await fetchConversations()
    } catch (err) {
      errorMsg.value = err.response?.data?.error || err.toString()
    }
  }
  
  // Fetch conversations when the component is mounted.
  onMounted(() => {
    fetchConversations()
  })
  </script>
  
  <style scoped>
  /* Add styles for ChatListView as needed */
  </style>
  