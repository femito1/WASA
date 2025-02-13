<!-- File: webui/src/components/GroupSettingsModal.vue -->
<template>
  <div class="modal-overlay">
    <div class="modal-content">
      <h5>Conversation Settings</h5>
      <div class="mb-3">
        <label for="groupName" class="form-label">Conversation Name:</label>
        <input id="groupName" v-model="groupName" type="text" class="form-control" />
      </div>
      <div class="mb-3">
        <label for="groupPhotoUpload" class="form-label">Upload Conversation Photo:</label>
        <input id="groupPhotoUpload" type="file" accept="image/*" @change="handleFileChange" class="form-control" />
      </div>
      <div class="mb-3">
        <label for="userSelect" class="form-label">Add Member:</label>
        <select id="userSelect" v-model="selectedUserId" class="form-select">
          <option v-for="u in users" :key="u.id" :value="u.id">
            {{ u.username }}
          </option>
        </select>
        <button class="btn btn-secondary mt-2" @click="addMember" :disabled="addingMember">
          Add Member
        </button>
      </div>
      <div class="d-flex justify-content-between">
        <button class="btn btn-danger" @click="leaveConversation" :disabled="processing">Leave Conversation</button>
        <div>
          <button class="btn btn-secondary me-2" @click="close">Close</button>
          <button class="btn btn-primary" @click="updateConversation" :disabled="processing">Update</button>
        </div>
      </div>
      <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '../services/axios'
import ErrorMsg from './ErrorMsg.vue'
import jwtDecode from 'jwt-decode'
import { useRouter } from 'vue-router'

const props = defineProps({
  conversation: {
    type: Object,
    required: true,
  },
})
const emit = defineEmits(['updated', 'close'])

const groupName = ref(props.conversation.name || '')
const processing = ref(false)
const addingMember = ref(false)
const errorMsg = ref(null)

const users = ref([])
const selectedUserId = ref(null)

const token = localStorage.getItem("authToken")
if (!token) {
  throw new Error("No authentication token found")
}
const decoded = jwtDecode(token)
const userId = Number(decoded.user_id)
const convId = props.conversation.id

const router = useRouter()

// Removed groupPhoto; using selectedFile instead
const selectedFile = ref(null)

async function fetchUsers() {
  try {
    const response = await axios.get(`/users`)
    users.value = response.data
    if (users.value.length > 0) {
      selectedUserId.value = users.value[0].id
    }
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
}

onMounted(() => {
  fetchUsers()
})

async function updateConversation() {
  processing.value = true
  errorMsg.value = null
  try {
    await axios.put(`/users/${userId}/conversations/${convId}/name`, { newName: groupName.value })
    if (selectedFile.value) {
      const base64Data = await fileToBase64(selectedFile.value)
      await axios.put(`/users/${userId}/conversations/${convId}/photo`, { newPhoto: base64Data })
    }
    emit("updated")
    close()
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
  processing.value = false
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

async function addMember() {
  if (!selectedUserId.value) return
  addingMember.value = true
  errorMsg.value = null
  try {
    await axios.post(`/users/${userId}/conversations/${convId}/members`, { userIdToAdd: selectedUserId.value })
    emit("updated")
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
  addingMember.value = false
}

async function leaveConversation() {
  if (!confirm("Are you sure you want to leave this conversation?")) return
  
  processing.value = true
  errorMsg.value = null
  
  try {
    await axios.delete(`/users/${userId}/conversations/${convId}/members`)
    emit("updated")
    router.push('/chats') // Redirect to chat list after leaving
  } catch (err) {
    errorMsg.value = err.response?.data?.error || 
      "You cannot leave this conversation. You might be the last member or this is a direct conversation."
  }
  
  processing.value = false
}

function close() {
  emit("close")
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
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
