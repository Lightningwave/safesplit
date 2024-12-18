const API_BASE_URL = 'http://localhost:8080/api';

export const loginSuperAdmin = async (email, password) => {
    try {
        const response = await fetch(`${API_BASE_URL}/super-login`, { 
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });

        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || 'Authentication failed');
        }

        if (data.data.user.role !== 'super_admin') {
            throw new Error('Invalid super admin credentials');
        }

        localStorage.setItem('token', data.token);
        localStorage.setItem('user', JSON.stringify(data.data.user));
        localStorage.setItem('billing', JSON.stringify(data.data.billing_profile));
        
        return data.data;
    } catch (error) {
        console.error('Super admin login error:', error);
        throw error;
    }
};

export const login = async (email, password) => {
    try {
        const response = await fetch(`${API_BASE_URL}/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });

        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || 'Login failed');
        }

        // Prevent super admin login through regular login
        if (data.data.user.role === 'super_admin') {
            throw new Error('Please use super admin login page');
        }

        localStorage.setItem('token', data.token);
        localStorage.setItem('user', JSON.stringify(data.data.user));
        if (data.data.billing_profile) {
            localStorage.setItem('billing', JSON.stringify(data.data.billing_profile));
        }
        
        return data.data;
    } catch (error) {
        console.error('Login error:', error);
        throw error;
    }
};

export const register = async (username, email, password) => {
    try {
        const response = await fetch(`${API_BASE_URL}/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username,
                email,
                password
            }),
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Registration failed');
        }

        return data;
    } catch (error) {
        console.error('Registration error:', error);
        throw error;
    }
};
export const getCurrentBilling = () => {
    const billing = localStorage.getItem('billing');
    return billing ? JSON.parse(billing) : null;
};
export const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    localStorage.removeItem('billing');
};

export const getCurrentUser = () => {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
};

export const getDashboardByRole = (role) => {
    switch (role) {
        case 'end_user':
            return '/dashboard';
        case 'premium_user':
            return '/premium-dashboard';
        case 'sys_admin':
            return '/admin-dashboard';
        case 'super_admin':
            return '/super-dashboard';
        default:
            return '/dashboard';
    }
};