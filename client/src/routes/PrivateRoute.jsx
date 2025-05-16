import { Navigate } from "react-router-dom";

const PrivateRoute = ({ children }) => {
    const sessionId = localStorage.getItem("session_id");

    if (!sessionId) {
        return <Navigate to="/auth" replace />;
    }

    return children;
};

export default PrivateRoute;
