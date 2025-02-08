<!-- File: webui/src/views/ChatView.vue -->
<template>
  <div class="container mt-3">
    <h1>Chat Conversation</h1>
    <button v-if="conversation && conversation.members && conversation.members.length >= 1" class="btn btn-sm btn-outline-secondary ms-2" @click="openGroupSettings">
      Conversation Settings
    </button>
    <div class="chat-window border p-3 mb-3" style="height: 400px; overflow-y: auto;">
      <LoadingSpinner :loading="loading">
        <div v-if="!loading">
          <div v-for="message in messages" :key="message.id">
            <!-- Pass currentUserId prop to MessageItem -->
            <MessageItem
              :message="message"
              :currentUserId="userId"
              @react="handleReact"
              @forward="handleForward"
              @comment="handleComment"
              @deleteMessage="handleDeleteMessage"
              @deleteComment="handleDeleteComment"
            />
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

    <!-- Reaction Picker Modal -->
    <ReactionPicker
      v-if="showReactionPicker"
      :message="selectedMessage"
      @react="submitReaction"
      @close="showReactionPicker = false"
    />

    <!-- Forward Message Modal -->
    <ForwardMessageModal
      v-if="showForwardModal"
      :message="selectedMessage"
      @forward="submitForward"
      @close="showForwardModal = false"
    />

    <!-- Group Settings Modal (for group conversations) -->
    <GroupSettingsModal
      v-if="showGroupSettings"
      :conversation="conversation"
      @updated="refreshConversation"
      @close="showGroupSettings = false"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import axios from '../services/axios'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import ErrorMsg from '../components/ErrorMsg.vue'
import MessageItem from '../components/MessageItem.vue'
import ReactionPicker from '../components/ReactionPicker.vue'
import ForwardMessageModal from '../components/ForwardMessageModal.vue'
import GroupSettingsModal from '../components/GroupSettingsModal.vue'
import jwtDecode from 'jwt-decode'

const route = useRoute()
const convId = route.params.convId
const selectedFile = ref(null)
const messages = ref([])
const newMessage = ref('')
const loading = ref(false)
const sending = ref(false)
const errorMsg = ref(null)

// Modal state
const showReactionPicker = ref(false)
const showForwardModal = ref(false)
const showGroupSettings = ref(false)
const selectedMessage = ref(null)
const conversation = ref(null)

const token = localStorage.getItem('authToken')
if (!token) {
  throw new Error('No authentication token found')
}
const decodedToken = jwtDecode(token)
const userId = Number(decodedToken.user_id)

async function fetchMessages() {
  loading.value = true
  errorMsg.value = null
  try {
    // API endpoint: GET /users/:id/conversations/:convId
    const response = await axios.get(`/users/${userId}/conversations/${convId}`)
    conversation.value = response.data  // Save entire conversation (may include updated members, name, photo)
    messages.value = response.data.messages || []
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
  loading.value = false
}

function handleFileChange(e) {
  if (e.target.files && e.target.files.length > 0) {
    selectedFile.value = e.target.files[0]
  } else {
    selectedFile.value = null
  }
}

async function sendMessage() {
  if (selectedFile.value) {
    sending.value = true
    errorMsg.value = null
    const reader = new FileReader()
    reader.onload = async (e) => {
      const base64Image = e.target.result
      try {
        // Send image message (format "image")
        const payload = { content: base64Image, format: "image" }
        const response = await axios.post(`/users/${userId}/conversations/${convId}/messages`, payload)
        messages.value.push(response.data)
        selectedFile.value = null
        newMessage.value = ''
      } catch (err) {
        errorMsg.value = err.response?.data?.error || err.toString()
      }
      sending.value = false
    }
    reader.onerror = () => {
      errorMsg.value = "Failed to read image file."
      sending.value = false
    }
    reader.readAsDataURL(selectedFile.value)
  } else if (newMessage.value.trim()) {
    sending.value = true
    errorMsg.value = null
    try {
      const payload = { content: newMessage.value, format: "string" }
      const response = await axios.post(`/users/${userId}/conversations/${convId}/messages`, payload)
      messages.value.push(response.data)
      newMessage.value = ''
    } catch (err) {
      errorMsg.value = err.response?.data?.error || err.toString()
    }
    sending.value = false
  }
}

// Handle reaction: open ReactionPicker modal for the selected message.
function handleReact(message) {
  selectedMessage.value = message
  showReactionPicker.value = true
}

// Submit reaction (POST /users/:id/conversations/:convId/messages/:msgId/reaction)
async function submitReaction(emoji) {
  try {
    await axios.post(`/users/${userId}/conversations/${convId}/messages/${selectedMessage.value.id}/reaction`, { emoji })
    // Refresh messages after reacting.
    await fetchMessages()
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
  showReactionPicker.value = false
  selectedMessage.value = null
}

// Handle forward: open ForwardMessageModal.
function handleForward(message) {
  selectedMessage.value = message
  showForwardModal.value = true
}

// Submit forward (POST /users/:id/conversations/:convId/messages/:msgId/forward)
async function submitForward(targetConversationId) {
  try {
    await axios.post(`/users/${userId}/conversations/${convId}/messages/${selectedMessage.value.id}/forward`, { targetConversationId })
    // Optionally, you might show a success message.
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
  showForwardModal.value = false
  selectedMessage.value = null
}

// Handle comment: prompt for comment text and POST comment.
async function handleComment(message) {
  const commentText = prompt("Enter your comment:")
  if (!commentText) return
  try {
    await axios.post(`/users/${userId}/conversations/${convId}/messages/${message.id}/comment`, { commentText })
    await fetchMessages()
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
}

// Handle delete message.
async function handleDeleteMessage(message) {
  if (!confirm("Are you sure you want to delete this message?")) return
  try {
    await axios.delete(`/users/${userId}/conversations/${convId}/messages/${message.id}`)
    await fetchMessages()
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
}

// Handle delete comment.
async function handleDeleteComment(message, commentId) {
  if (!confirm("Are you sure you want to delete this comment?")) return
  try {
    await axios.delete(`/users/${userId}/conversations/${convId}/messages/${message.id}/comment/${commentId}`)
    await fetchMessages()
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
}

// For group conversations: open group settings if conversation is a group (more than 2 members)
function openGroupSettings() {
  showGroupSettings.value = true
}

async function refreshConversation() {
  await fetchMessages()
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
