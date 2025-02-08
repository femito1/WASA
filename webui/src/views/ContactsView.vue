<!-- File: webui/src/views/ContactsView.vue -->
<template>
  <div class="container mt-3">
    <h1>Your Contacts</h1>
    <LoadingSpinner :loading="loading">
      <div v-if="!loading">
        <ul v-if="contacts.length" class="list-group">
          <li v-for="contact in contacts" :key="contact.id" class="list-group-item">
            {{ contact.username }}
            <button class="btn btn-sm btn-danger float-end" @click="remove(contact.id)">Remove</button>
          </li>
        </ul>
      </div>
    </LoadingSpinner>
    <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
    <div class="mt-3">
      <input v-model="searchTerm" type="text" placeholder="Search users..." class="form-control mb-2" />
      <button class="btn btn-primary" @click="searchUsers">Search and Add Contact</button>
    </div>
    <ul v-if="searchedUsers.length" class="list-group mt-2">
      <li v-for="user in searchedUsers" :key="user.id" class="list-group-item">
        {{ user.username }}
        <button class="btn btn-sm btn-success float-end" @click="add(user.id)">Add Contact</button>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '../services/axios'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import ErrorMsg from '../components/ErrorMsg.vue'
import jwtDecode from 'jwt-decode'
import { useRouter } from 'vue-router'

const contacts = ref([])
const loading = ref(false)
const errorMsg = ref(null)
const searchTerm = ref('')
const searchedUsers = ref([])
const router = useRouter()

const token = localStorage.getItem('authToken')
if (!token) {
  router.push({ name: 'Login' })
  throw new Error('No authentication token found.')
}
const decoded = jwtDecode(token)
const userId = decoded.user_id

async function fetchContacts() {
  loading.value = true
  errorMsg.value = null
  try {
    const response = await axios.get(`/users/${userId}/contacts`)
    contacts.value = response.data
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
  loading.value = false
}

async function searchUsers() {
  // Use GET /users?name=searchTerm to search for users
  try {
    const response = await axios.get(`/users?name=${searchTerm.value}`)
    searchedUsers.value = response.data
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
}

async function add(contactId) {
  try {
    await axios.post(`/users/${userId}/contacts`, { contactId })
    await fetchContacts()
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
}

async function remove(contactId) {
  try {
    await axios.delete(`/users/${userId}/contacts/${contactId}`)
    await fetchContacts()
  } catch (err) {
    errorMsg.value = err.response?.data?.error || err.toString()
  }
}

onMounted(() => {
  fetchContacts()
})
</script>
