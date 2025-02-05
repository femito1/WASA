<!-- File: webui/src/App.vue -->
<template>
	<div>
	  <!-- Header -->
	  <header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
		<div class="container-fluid">
		  <RouterLink to="/" class="navbar-brand px-3 fs-6">WhatsApp Clone</RouterLink>
		  <nav class="d-flex align-items-center">
			<RouterLink to="/chats" class="nav-link text-white me-2">Chats</RouterLink>
			<RouterLink to="/contacts" class="nav-link text-white me-2">Contacts</RouterLink>
			<div class="ms-auto d-flex align-items-center">
			  <template v-if="isAuthenticated">
				<span class="text-white me-2">Hello, {{ username }}</span>
				<button class="btn btn-sm btn-outline-light" @click="logout">Logout</button>
			  </template>
			  <template v-else>
				<RouterLink to="/" class="btn btn-sm btn-outline-light">Login</RouterLink>
			  </template>
			</div>
		  </nav>
		  <button class="navbar-toggler d-md-none collapsed" type="button" data-bs-toggle="collapse"
				  data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false"
				  aria-label="Toggle navigation">
			<span class="navbar-toggler-icon"></span>
		  </button>
		</div>
	  </header>
  
	  <!-- Layout with Sidebar and Main Content -->
	  <div class="container-fluid">
		<div class="row">
		  <nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
			<div class="position-sticky pt-3">
			  <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
				<span>Menu</span>
			  </h6>
			  <ul class="nav flex-column">
				<li class="nav-item">
				  <RouterLink to="/chats" class="nav-link">
					<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#message-circle" /></svg>
					Chats
				  </RouterLink>
				</li>
				<li class="nav-item">
				  <RouterLink to="/contacts" class="nav-link">
					<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#users" /></svg>
					Contacts
				  </RouterLink>
				</li>
			  </ul>
			</div>
		  </nav>
  
		  <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
			<RouterView />
		  </main>
		</div>
	  </div>
	</div>
  </template>
  
  <script setup>
  import { ref, computed } from 'vue'
  import { useRouter, RouterLink, RouterView } from 'vue-router'
  import jwtDecode from 'jwt-decode'
  
  // Authentication state via localStorage.
  const token = ref(localStorage.getItem('authToken') || null)
  const decoded = computed(() => {
	if (token.value) {
	  try {
		return jwtDecode(token.value)
	  } catch (e) {
		console.error('Token decode error:', e)
		return null
	  }
	}
	return null
  })
  const username = computed(() => {
	return decoded.value && decoded.value.username ? decoded.value.username : 'User'
  })
  const isAuthenticated = computed(() => !!token.value)
  const router = useRouter()
  
  function logout() {
	token.value = null
	localStorage.removeItem('authToken')
	localStorage.removeItem('userId')
	localStorage.removeItem('username')
	router.push({ name: 'Login' })
  }
  </script>
  
  <style scoped>
  /* Add App.vue specific styles if needed */
  </style>
  