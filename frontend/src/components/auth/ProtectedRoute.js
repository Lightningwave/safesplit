import { Navigate } from 'react-router-dom';
import { getDashboardByRole, getCurrentUser } from '../../services/authService';

const ProtectedRoute = ({ element: Element, allowedRoles }) => {
    const user = getCurrentUser();
    
    if (!user) {
        return <Navigate to="/login" replace />;
    }

    if (allowedRoles && !allowedRoles.includes(user.role)) {
        const currentDashboard = getDashboardByRole(user.role);
        return <Navigate to={currentDashboard} replace />;
    }

    return <Element user={user} />;
};

export default ProtectedRoute;