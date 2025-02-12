import jwtDecode from 'jwt-decode'

class WebSocketService {
  constructor() {
    this.ws = null
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
    this.listeners = new Map()
    this.currentConversationId = null
    this.reconnectTimeout = null
    this.pendingMessages = new Set() // Track messages waiting for delivery confirmation
    this.connected = false
  }

  connect() {
    if (this.connected) return
    if (this.ws?.readyState === WebSocket.OPEN) return

    const token = localStorage.getItem('authToken')
    if (!token) return

    const decoded = jwtDecode(token)
    const userId = decoded.user_id

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${wsProtocol}//${window.location.host}/users/${userId}/ws`

    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      console.log('WebSocket connected')
      this.connected = true
      this.reconnectAttempts = 0
      if (this.currentConversationId) {
        this.subscribeToConversation(this.currentConversationId)
      }
      this.notifyConnected()
    }

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        const listeners = this.listeners.get(data.type) || []
        listeners.forEach(callback => callback(data.payload))
      } catch (err) {
        console.error('WebSocket message error:', err)
      }
    }

    this.ws.onclose = () => {
      this.connected = false
      this.retryConnection()
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }
  }

  retryConnection() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectTimeout = setTimeout(() => {
        this.reconnectAttempts++
        this.connect()
      }, Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000))
    }
  }

  handleMessage(event) {
    try {
      const data = JSON.parse(event.data)
      const listeners = this.listeners.get(data.type) || []
      
      if (!data.type || !data.payload) {
        console.error('Invalid message format:', data)
        return
      }

      switch (data.type) {
        case 'new_message':
        case 'message_update':
        case 'message_read':
        case 'message_reaction':
          listeners.forEach(callback => callback(data.payload))
          break
          
        case 'profile_update':
          this.handleProfileUpdate(data.payload)
          break
          
        default:
          console.warn('Unknown message type:', data.type)
      }
    } catch (err) {
      console.error('WebSocket message error:', err)
      this.retryConnection()
    }
  }

  handleProfileUpdate(payload) {
    const listeners = this.listeners.get('profile_update') || []
    listeners.forEach(callback => callback(payload))
    
    // Also update local storage if it's the current user
    const token = localStorage.getItem('authToken')
    if (token) {
      const decoded = jwtDecode(token)
      if (decoded.user_id === payload.userId) {
        // Update profile picture in local storage or relevant state management
        this.emit('profile_updated', payload.profilePicture)
      }
    }
  }

  subscribeToConversation(conversationId) {
    if (!conversationId) return
    
    this.currentConversationId = conversationId
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: 'subscribe',
        payload: conversationId
      }))
    }
  }

  subscribe(type, callback) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, [])
    }
    this.listeners.get(type).push(callback)
  }

  unsubscribe(type, callback) {
    const listeners = this.listeners.get(type)
    if (listeners) {
      this.listeners.set(type, listeners.filter(cb => cb !== callback))
    }
  }

  disconnect() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.currentConversationId = null
    this.listeners.clear()
  }

  async resendPendingMessages() {
    for (const msgId of this.pendingMessages) {
      // Attempt to resend message status updates
      try {
        await this.sendMessageStatus(msgId, 'sent')
      } catch (err) {
        console.error('Failed to resend message status:', err)
      }
    }
  }

  notifyConnected() {
    if (this.currentConversationId) {
      const token = localStorage.getItem('authToken')
      if (token) {
        this.ws.send(JSON.stringify({
          type: 'connected',
          payload: { userId: jwtDecode(token).user_id }
        }))
      }
    }
  }

  handleMessageUpdate(update) {
    const listeners = this.listeners.get('message_update') || []
    listeners.forEach(callback => {
      try {
        callback(update)
      } catch (err) {
        console.error('Error in message update listener:', err)
      }
    })
  }
}

export default new WebSocketService() 