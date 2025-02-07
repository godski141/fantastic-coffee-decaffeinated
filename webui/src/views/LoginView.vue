<script>

export default {
  data() {
    return {
      username: "",  // Nome utente inserito nel form
      errorMsg: null, // Messaggio di errore se il login fallisce
      loading: false, // Indica se la richiesta è in corso
      userId: null    // ID utente restituito dal backend
    };
  },
  methods: {
    async login() {
      this.errorMsg = null;
      this.loading = true;

      // Validazione lato frontend
      if (this.username.length < 3 || this.username.length > 50) {
        this.errorMsg = "Username must be between 3 and 50 characters.";
        this.loading = false;
        return;
      }

      try {
        // Chiamata API al backend
        const response = await this.$axios.post('/session', { username: this.username });

        // Salva l'ID utente restituito
        this.userId = response.data.user_id;
        sessionStorage.setItem('authToken', `Bearer ${this.userId}`); // Salva il token
        sessionStorage.setItem('username', this.username);

        // Reindirizza alla dashboard o a un'altra pagina
        this.$router.push('/dashboard');
      } catch (error) {
        // Gestione degli errori
        this.errorMsg = error.response?.data || "Login failed. Please try again.";
      } finally {
        this.loading = false;
      }
    }
  }
};
</script>

<template>
  <div class="session-container">
    <h1>Login</h1>

    <!-- Messaggio di errore -->
    <p v-if="errorMsg" class="error">{{ errorMsg }}</p>

    <!-- Form di login -->
    <form @submit.prevent="login">
      <label for="username">Username:</label>
      <input 
        type="text" 
        id="username" 
        v-model="username" 
        required 
        minlength="3" 
        maxlength="50"
      />
      <button type="submit" :disabled="loading">
        {{ loading ? "Logging in..." : "Login" }}
      </button>
    </form>

    <!-- Messaggio di benvenuto dopo il login -->
    <p v-if="userId">Welcome, {{ username }}!</p>
  </div>
</template>

<style>
/* Stili base */
.session-container {
  max-width: 400px;
  margin: 0 auto;
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #f9f9f9;
}

input {
  width: 100%;
  padding: 8px;
  margin: 10px 0;
}

button {
  padding: 10px;
  background-color: #007bff;
  color: white;
  border: none;
  cursor: pointer;
}

button:disabled {
  background-color: #aaa;
}

.error {
  color: red;
}
</style>
