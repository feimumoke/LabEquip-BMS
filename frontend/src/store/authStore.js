import { makeAutoObservable } from 'mobx';
import { getCurrentUser } from '../utils/auth';

class AuthStore {
  user = null;
  isAuthenticated = false;

  constructor() {
    makeAutoObservable(this);
    this.init();
  }

  init() {
    const user = getCurrentUser();
    if (user) {
      this.user = user;
      this.isAuthenticated = true;
    }
  }

  setUser(user) {
    this.user = user;
    this.isAuthenticated = !!user;
  }

  logout() {
    this.user = null;
    this.isAuthenticated = false;
    localStorage.removeItem('token');
    localStorage.removeItem('userEmail');
    localStorage.removeItem('userInfo');
  }

  get isAdmin() {
    return this.user && [1, 2, 3].includes(this.user.role);
  }

  get isTeacher() {
    return this.user && [1, 2, 3].includes(this.user.role);
  }

  get isSuperAdmin() {
    return this.user && this.user.role === 1;
  }
}

export default new AuthStore();

