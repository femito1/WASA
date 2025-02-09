import axios from "axios";

const instance = axios.create({
  baseURL: __API_URL__,
  timeout: 5000, // 5-second timeout
});

instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("authToken");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

instance.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error("Axios response error:", error);
    return Promise.reject(error);
  }
);

export default instance;
