<!-- File: webui/src/views/ChatView.vue -->
<template>
    <div class="container mt-3">
      <h1>Chat Conversation</h1>
      <div class="chat-window border p-3 mb-3" style="height: 400px; overflow-y: auto;">
        <LoadingSpinner :loading="loading">
          <div v-if="!loading">
            <div v-for="message in messages" :key="message.id">
              <MessageItem :message="message" />
            </div>
          </div>
        </LoadingSpinner>
        <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
      </div>
      <div class="mb-3">
        <label for="imageUpload" class="form-label">Attach Image:</label>
        <input id="imageUpload" type="file" accept="image/*" @change="handleFileChange" class="form-control">
      </div>
      <form @submit.prevent="sendMessage">
        <div class="input-group">
          <input v-model="newMessage" type="text" class="form-control" placeholder="Type your message..." :disabled="sending || selectedFile">
          <button class="btn btn-primary" type="submit" :disabled="sending">
            <span v-if="sending">Sending...</span>
            <span v-else>Send</span>
        </button>
        </div>
      </form>
    </div>
  </template>
  
  <script setup>
  import { ref, onMounted } from 'vue'
  import { useRoute } from 'vue-router'
  import axios from '../services/axios'
  import LoadingSpinner from '../components/LoadingSpinner.vue'
  import ErrorMsg from '../components/ErrorMsg.vue'
  import MessageItem from '../components/MessageItem.vue'
  
  const route = useRoute()
  const convId = route.params.convId
  const selectedFile = ref(null)
  const messages = ref([])
  const newMessage = ref('')
  const loading = ref(false)
  const sending = ref(false)
  const errorMsg = ref(null)
  
  import jwtDecode from 'jwt-decode'

  const token = localStorage.getItem('authToken')
  if (!token) {
  // Optionally, redirect to login
  throw new Error('No authentication token found')
  }
  const decodedToken = jwtDecode(token)
  const userId = decodedToken.user_id  // Adjust this field name if your token uses a different key.


  async function fetchMessages() {
    loading.value = true
    errorMsg.value = null
    try {
      // API endpoint: GET /users/:id/conversations/:convId
      const response = await axios.get(`/users/${userId}/conversations/${convId}`)
      // Assume the conversation object has a messages field.
      messages.value = response.data.messages || []
    } catch (err) {
      errorMsg.value = err.response?.data?.error || err.toString()
    }
    loading.value = false
  }
  
function handleFileChange(e) {
  if (e.target.files && e.target.files.length > 0) {
    selectedFile.value = e.target.files[0];
  } else {
    selectedFile.value = null;
  }
}

async function sendMessage() {
  // If an image file is selected, process it as an image message.
  if (selectedFile.value) {
    sending.value = true;
    errorMsg.value = null;
    const reader = new FileReader();
    reader.onload = async (e) => {
      const base64Image = e.target.result;
      try {
        // Send the message with format "image"
        const payload = { content: base64Image, format: "image" };
        const response = await axios.post(`/users/${userId}/conversations/${convId}/messages`, payload);
        messages.value.push(response.data);
        // Clear both file and any text message.
        selectedFile.value = null;
        newMessage.value = '';
      } catch (err) {
        errorMsg.value = err.response?.data?.error || err.toString();
      }
      sending.value = false;
    };
    reader.onerror = () => {
      errorMsg.value = "Failed to read image file.";
      sending.value = false;
    };
    reader.readAsDataURL(selectedFile.value);
  } else if (newMessage.value.trim()) {
    // Otherwise, send a text message.
    sending.value = true;
    errorMsg.value = null;
    try {
      const payload = { content: newMessage.value, format: "string" };
      const response = await axios.post(`/users/${userId}/conversations/${convId}/messages`, payload);
      messages.value.push(response.data);
      newMessage.value = '';
    } catch (err) {
      errorMsg.value = err.response?.data?.error || err.toString();
    }
    sending.value = false;
  }
}
  
  onMounted(() => {
    fetchMessages()
  })
  </script>
  
  <style scoped>
  .chat-window {
    background-color: #f8f9fa;
  }
  </style>
  