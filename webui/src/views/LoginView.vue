<script>
import { doLogin, saveUserData } from '../services/axios.js'

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
        // Chiamata API al backend usando la nostra funzione
        const response = await doLogin(this.username);

        // Salva i dati utente usando la nostra funzione
        this.userId = response.user_id;
        saveUserData(this.userId, this.username);

        // Reindirizza alla home dell'app
        this.$router.push('/home');
      } catch (error) {
        // Gestione degli errori
        console.error('Login error:', error);
        this.errorMsg = error.response?.data || "Login failed. Please try again.";
      } finally {
        this.loading = false;
      }
    }
  }
};
</script>

<template>
  <div class="login-container">
    <div class="login-card">
      <!-- Logo/Header -->
      <div class="login-header">
        <h1>WasaTEXT</h1>
        <p>Messaggistica Istantanea</p>
      </div>

      <!-- Messaggio di errore -->
      <div v-if="errorMsg" class="alert alert-danger" role="alert">
        {{ errorMsg }}
      </div>

      <!-- Form di login -->
      <form @submit.prevent="login" class="login-form">
        <div class="form-group">
          <label for="username" class="form-label">Username</label>
          <input 
            type="text" 
            id="username" 
            v-model="username" 
            class="form-control"
            placeholder="Inserisci il tuo username"
            required 
            minlength="3" 
            maxlength="50"
            :disabled="loading"
          />
          <div class="form-text">Username deve essere tra 3 e 50 caratteri</div>
        </div>
        
        <button 
          type="submit" 
          class="btn btn-primary w-100"
          :disabled="loading || username.length < 3"
        >
          <span v-if="loading" class="spinner-border spinner-border-sm me-2" role="status"></span>
          {{ loading ? "Accesso in corso..." : "Accedi" }}
        </button>
      </form>

      <!-- Messaggio di benvenuto dopo il login -->
      <div v-if="userId" class="alert alert-success mt-3" role="alert">
        Benvenuto, {{ username }}! Reindirizzamento in corso...
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Container principale */
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

/* Card di login */
.login-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  padding: 2rem;
  width: 100%;
  max-width: 400px;
  animation: slideUp 0.5s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Header */
.login-header {
  text-align: center;
  margin-bottom: 2rem;
}

.login-header h1 {
  color: #333;
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.login-header p {
  color: #666;
  font-size: 1rem;
  margin: 0;
}

/* Form */
.login-form {
  margin-bottom: 1rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #333;
}

.form-control {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e1e5e9;
  border-radius: 8px;
  font-size: 1rem;
  transition: all 0.3s ease;
}

.form-control:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-control:disabled {
  background-color: #f8f9fa;
  cursor: not-allowed;
}

.form-text {
  font-size: 0.875rem;
  color: #6c757d;
  margin-top: 0.25rem;
}

/* Button */
.btn {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.3);
}

.btn-primary:disabled {
  background: #6c757d;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.w-100 {
  width: 100%;
}

/* Alerts */
.alert {
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 1rem;
  border: none;
}

.alert-danger {
  background-color: #f8d7da;
  color: #721c24;
  border-left: 4px solid #dc3545;
}

.alert-success {
  background-color: #d1edff;
  color: #0c5460;
  border-left: 4px solid #0dcaf0;
}

/* Spinner */
.spinner-border-sm {
  width: 1rem;
  height: 1rem;
}

.me-2 {
  margin-right: 0.5rem;
}

.mt-3 {
  margin-top: 1rem;
}

/* Responsive */
@media (max-width: 480px) {
  .login-container {
    padding: 10px;
  }
  
  .login-card {
    padding: 1.5rem;
  }
  
  .login-header h1 {
    font-size: 2rem;
  }
}
</style>
