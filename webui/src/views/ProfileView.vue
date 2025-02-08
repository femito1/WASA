<!-- File: webui/src/views/ProfileView.vue -->
<template>
    <div class="container mt-3">
      <h1>Profile</h1>
      <form @submit.prevent="updateProfile">
        <div class="mb-3">
          <label for="username" class="form-label">Username:</label>
          <input id="username" v-model="newUsername" type="text" class="form-control" required />
        </div>
        <div class="mb-3">
          <label for="photo" class="form-label">Profile Picture (Base64):</label>
          <textarea id="photo" v-model="newPhoto" class="form-control" rows="3" placeholder="Paste Base64 encoded image"></textarea>
        </div>
        <button type="submit" class="btn btn-primary" :disabled="updating">
          <span v-if="updating">Updating...</span>
          <span v-else>Update Profile</span>
        </button>
      </form>
      <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
    </div>
  </template>
  
  <script setup>
  import { ref, onMounted } from 'vue'
  import axios from '../services/axios'
  import ErrorMsg from '../components/ErrorMsg.vue'
  import jwtDecode from 'jwt-decode'
  import { useRouter } from 'vue-router'
  
  const router = useRouter()
  const newUsername = ref('')
  const newPhoto = ref('')
  const updating = ref(false)
  const errorMsg = ref(null)
  
  const token = localStorage.getItem("authToken")
  if (!token) {
    router.push({ name: 'Login' })
    throw new Error("No authentication token found")
  }
  const decoded = jwtDecode(token)
  const userId = Number(decoded.user_id)
  
  // Optionally, fetch current user info to prefill the form (if your backend supports it).
  // Here we assume the token contains the username.
  newUsername.value = decoded.username || ''
  
  async function updateProfile() {
    updating.value = true
    errorMsg.value = null
    try {
      // Update username (PUT /users/:id)
      await axios.put(`/users/${userId}`, { newName: newUsername.value })
      // Update profile picture (PUT /users/:id/photo) if provided
      if (newPhoto.value.trim()) {
        await axios.put(`/users/${userId}/photo`, { newPic: newPhoto.value })
      }
      alert("Profile updated successfully!")
    } catch (err) {
      errorMsg.value = err.response?.data?.error || err.toString()
    }
    updating.value = false
  }
  </script>
  
  <style scoped>
  .container {
    max-width: 500px;
  }
  </style>
  