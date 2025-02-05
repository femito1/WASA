<!-- File: webui/src/views/LoginView.vue -->
<template>
    <div class="container mt-5">
      <h1>Login</h1>
      <form @submit.prevent="login">
        <div class="mb-3">
          <label for="username" class="form-label">Username:</label>
          <input id="username" v-model="username" type="text" class="form-control" required />
        </div>
        <button type="submit" class="btn btn-primary" :disabled="loading">
          <span v-if="loading">Logging in...</span>
          <span v-else>Login</span>
        </button>
      </form>
      <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
    </div>
  </template>
  
  <script setup>
  import { ref } from 'vue'
  import { useRouter } from 'vue-router'
  import axios from '../services/axios'
  import ErrorMsg from '../components/ErrorMsg.vue'
  
  const username = ref('')
  const loading = ref(false)
  const errorMsg = ref(null)
  const router = useRouter()
  
  async function login() {
    loading.value = true
    errorMsg.value = null
    try {
      // The API specification for /session expects { name: string }
      const response = await axios.post('/session', { name: username.value })
      // The API returns { identifier: token }
      const token = response.data.identifier
      // Save the token
      localStorage.setItem('authToken', token)
      // Navigate to the chat list
      router.push({ name: 'ChatList' })
    } catch (err) {
      errorMsg.value = err.response?.data?.error || err.toString()
    }
    loading.value = false
  }
  </script>
  
  <style scoped>
  /* Basic styling for the login view */
  .container {
    max-width: 400px;
  }
  </style>
  