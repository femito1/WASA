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
          <label for="photoUpload" class="form-label">Upload Profile Picture:</label>
          <input id="photoUpload" type="file" accept="image/*" @change="handleFileChange" class="form-control" />
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
  const selectedFile = ref(null)
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
      // Try to update username
      if (newUsername.value !== decoded.username) {
        try {
          await axios.put(`/users/${userId}`, { newName: newUsername.value })
        } catch (err) {
          if (err.response?.status === 500) {
            errorMsg.value = "This username is already taken. Please choose another one."
            updating.value = false
            return
          }
          throw err
        }
      }
      
      // Update profile picture if a new file was uploaded
      if (selectedFile.value) {
        const base64Data = await fileToBase64(selectedFile.value)
        await axios.put(`/users/${userId}/photo`, { newPic: base64Data })
      }
      
      alert("Profile updated successfully!")
    } catch (err) {
      errorMsg.value = err.response?.data?.error || err.toString()
    }
    
    updating.value = false
  }
  
  function handleFileChange(e) {
    if (e.target.files && e.target.files.length > 0) {
      selectedFile.value = e.target.files[0]
    } else {
      selectedFile.value = null
    }
  }
  
  function fileToBase64(file) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.readAsDataURL(file);
      reader.onload = () => resolve(reader.result);
      reader.onerror = error => reject(error);
    });
  }
  </script>
  
  <style scoped>
  .container {
    max-width: 500px;
  }
  </style>
  