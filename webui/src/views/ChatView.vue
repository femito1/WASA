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
          <div v-if="replyingTo" class="reply-preview alert alert-info">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                Replying to {{ replyingTo.senderName }}:
                <br>
                <small>{{ truncateContent(replyingTo.content) }}</small>
              </div>
              <button class="btn btn-sm btn-close" @click="replyingTo = null"></button>
            </div>
          </div>
          <div v-for="message in messages" :key="message.id">
            <MessageItem
              :message="message"
              :currentUserId="userId"
              @react="handleReact"
              @reply="handleReply"
              @forward="handleForward"
              @deleteMessage="handleDeleteMessage"
              @removeReaction="handleRemoveReaction"
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
        <input v-model="newMessage" type="text" class="form-control" :placeholder="replyingTo ? 'Type your reply...' : 'Type your message...'" :disabled="sending || selectedFile">
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

const replyingTo = ref(null)

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
  if (!newMessage.value.trim() && !selectedFile.value) return
  
  sending.value = true
  errorMsg.value = null
  
  try {
    let content, format
    
    if (selectedFile.value) {
      const base64 = await fileToBase64(selectedFile.value)
      content = base64
      format = 'image'
    } else {
      content = newMessage.value
      format = 'string'
    }

    const payload = {
      content,
      format,
      replyTo: replyingTo.value?.id // Include reply reference if replying
    }

    await axios.post(`/users/${userId}/conversations/${convId}/messages`, payload)
    await fetchMessages()
    
    newMessage.value = ''
    selectedFile.value = null
    replyingTo.value = null
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
  
  sending.value = false
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
    await axios.post(
      `/users/${userId}/conversations/${convId}/messages/${selectedMessage.value.id}/forward`,
      { targetConversationId }
    );

    // If forwarding within the same conversation, update the list so the forwarded message (with forwarded indicator)
    // is visible. Otherwise, notify the user (e.g., via a toast or alert).
    if (targetConversationId === convId) {
      await fetchMessages();
    } else {
      alert("Message forwarded successfully to the selected conversation.");
    }
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString();
  }
  showForwardModal.value = false;
  selectedMessage.value = null;
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

// For group conversations: open group settings if conversation is a group (more than 2 members)
function openGroupSettings() {
  showGroupSettings.value = true
}

async function refreshConversation() {
  await fetchMessages()
}

async function handleReply(message) {
  replyingTo.value = message
}

async function handleRemoveReaction(message, emoji) {
  try {
    await axios.delete(`/users/${userId}/conversations/${convId}/messages/${message.id}/reaction/${emoji}`)
    await fetchMessages()
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result)
    reader.onerror = error => reject(error)
  })
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