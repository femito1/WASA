const token = localStorage.getItem("authToken");
const wsUrl = typeof __WS_API_URL__ !== "undefined" ? __WS_API_URL__ : "ws://localhost:3000/ws";
const socket = new WebSocket(wsUrl + "?token=" + token);

socket.onopen = () => {
  console.log("WebSocket connected");
};

socket.onclose = () => {
  console.log("WebSocket disconnected");
};

socket.onerror = (error) => {
  console.error("WebSocket error:", error);
};

export default socket;
