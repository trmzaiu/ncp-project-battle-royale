import { useState } from "react";
import { useNavigate } from "react-router-dom";
import useWebSocket from "../hooks/useWebSocket";

export default function Auth() {
    const navigate = useNavigate();

    const [activeTab, setActiveTab] = useState("login");
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [newUsername, setNewUsername] = useState("");
    const [newPassword, setNewPassword] = useState("");
    const [loading, setLoading] = useState(false);
    const [notification, setNotification] = useState({ show: false, message: "", type: "info" });

    const showNotification = (message, type = "info") => {
        setNotification({ show: true, message, type });
        setTimeout(() => setNotification((prev) => ({ ...prev, show: false })), 4000);
    };

    const handleSocketMessage = (res) => {
        switch (res.type) {
            case "login_response":
                if (res.success) {
                    localStorage.setItem("session_id", res.data.session_id);
                    showNotification(res.message, "success");
                    setLoading(true);
                    setTimeout(() => navigate("/lobby"), 1000);
                } else {
                    showNotification(res.message, "error");
                }
                break;

            case "register_response":
                if (res.success) {
                    showNotification(res.message, "success");
                    setTimeout(() => setActiveTab("login"), 1500);
                } else {
                    showNotification(res.message, "error");
                }
                break;

            default:
                if (res.message) showNotification(res.message, "info");
                break;
        }
    };

    const { sendMessage } = useWebSocket(handleSocketMessage);

    const handleLogin = () => {
        if (!username || !password) return showNotification("Please fill in both fields.", "warning");
        sendMessage({ type: "login", data: { username, password } });
    };

    const handleRegister = () => {
        if (!newUsername || !newPassword) return showNotification("Please fill in all fields.", "warning");
        sendMessage({ type: "register", data: { username: newUsername, password: newPassword } });
    };

    return (
        <div className="min-h-screen flex items-center justify-center p-5 bg-gradient-to-br from-blue-500/80 to-orange-400/80 overflow-hidden">
            <div className="bg-white/90 backdrop-blur-md rounded-xl shadow-xl w-full max-w-md overflow-hidden relative transition-all duration-300 transform translate-z-0">
                <div className="flex flex-col items-center p-6 sm:p-8 bg-gradient-to-r from-white/95 to-white/85 border-b border-gray-200">
                    <h2 className="text-3xl font-bold mb-6">🏰 ROYAKA</h2>
                    <div className="flex w-full rounded-lg bg-black/4 overflow-hidden">
                        <button
                            onClick={() => setActiveTab("login")}
                            className={`flex-1 py-4 text-center font-semibold relative transition-all duration-300 ${activeTab === "login"
                                ? "text-blue-500 bg-blue-500/10 after:absolute after:bottom-0 after:left-0 after:w-full after:h-0.5 after:bg-blue-500"
                                : "text-gray-500 hover:text-blue-500"
                                }`}
                        >
                            Login
                        </button>
                        <button
                            onClick={() => setActiveTab("register")}
                            className={`flex-1 py-4 text-center font-semibold relative transition-all duration-300 ${activeTab === "register"
                                ? "text-blue-500 bg-blue-500/10 after:absolute after:bottom-0 after:left-0 after:w-full after:h-0.5 after:bg-blue-500"
                                : "text-gray-500 hover:text-blue-500"
                                }`}
                        >
                            Register
                        </button>
                    </div>
                </div>

                {activeTab === "login" ? (
                    <div className="p-6 sm:p-8 animate-fadeIn">
                        <div className="text-center mb-6">
                            <h2 className="text-2xl font-semibold text-gray-800 mb-2">Welcome Back!</h2>
                            <p className="text-gray-500 text-sm">Enter your credentials to access your account</p>
                        </div>

                        <div className="relative mb-5">
                            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500">
                                <i className="fas fa-user"></i>
                            </span>
                            <input
                                type="text"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                placeholder="Username"
                                className="w-full py-4 pl-12 pr-4 border-2 border-gray-200 rounded-lg text-base transition-all bg-gray-100 focus:border-blue-500 focus:bg-white focus:shadow-[0_0_0_4px_rgba(74,144,226,0.1)] outline-none"
                            />
                        </div>

                        <div className="relative mb-6">
                            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500">
                                <i className="fas fa-lock"></i>
                            </span>
                            <input
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                placeholder="Password"
                                className="w-full py-4 pl-12 pr-4 border-2 border-gray-200 rounded-lg text-base transition-all bg-gray-100 focus:border-blue-500 focus:bg-white focus:shadow-[0_0_0_4px_rgba(74,144,226,0.1)] outline-none"
                            />
                        </div>

                        <button
                            onClick={handleLogin}
                            className="w-full py-4 bg-gradient-to-r from-blue-500 to-blue-400 text-white rounded-lg font-semibold text-base cursor-pointer transition-all flex justify-center items-center gap-2 shadow-md shadow-blue-500/20 hover:shadow-lg hover:shadow-blue-500/30 hover:-translate-y-0.5 active:translate-y-0.5"
                        >
                            {loading ? (
                                <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                            ) : (
                                <>
                                    <span>Login</span>
                                    <i className="fas fa-arrow-right transition-transform duration-300 group-hover:translate-x-1"></i>
                                </>
                            )}
                        </button>
                    </div>
                ) : (
                    <div className="p-6 sm:p-8 animate-fadeIn">
                        <div className="text-center mb-6">
                            <h2 className="text-2xl font-semibold text-gray-800 mb-2">Create Account</h2>
                            <p className="text-gray-500 text-sm">Join the Royaka community today</p>
                        </div>

                        <div className="relative mb-5">
                            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500">
                                <i className="fas fa-user"></i>
                            </span>
                            <input
                                type="text"
                                value={newUsername}
                                onChange={(e) => setNewUsername(e.target.value)}
                                placeholder="Choose a username"
                                className="w-full py-4 pl-12 pr-4 border-2 border-gray-200 rounded-lg text-base transition-all bg-gray-100 focus:border-blue-500 focus:bg-white focus:shadow-[0_0_0_4px_rgba(74,144,226,0.1)] outline-none"
                            />
                        </div>

                        <div className="relative mb-6">
                            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500">
                                <i className="fas fa-lock"></i>
                            </span>
                            <input
                                type="password"
                                value={newPassword}
                                onChange={(e) => setNewPassword(e.target.value)}
                                placeholder="Create password"
                                className="w-full py-4 pl-12 pr-4 border-2 border-gray-200 rounded-lg text-base transition-all bg-gray-100 focus:border-blue-500 focus:bg-white focus:shadow-[0_0_0_4px_rgba(74,144,226,0.1)] outline-none"
                            />
                        </div>

                        <button
                            onClick={handleRegister}
                            className="w-full py-4 bg-gradient-to-r from-blue-500 to-blue-400 text-white rounded-lg font-semibold text-base cursor-pointer transition-all flex justify-center items-center gap-2 shadow-md shadow-blue-500/20 hover:shadow-lg hover:shadow-blue-500/30 hover:-translate-y-0.5 active:translate-y-0.5"
                        >
                            <span>Create Account</span>
                        </button>
                    </div>
                )}
            </div>

            <div
                className={`fixed top-5 right-5 flex items-center justify-center bg-white/15 backdrop-blur-md rounded-lg py-4 px-5 text-white text-sm shadow-xl border-l-4 z-50 transform transition-transform duration-300 ${notification.show ? "translate-x-0" : "translate-x-full"
                    } ${notification.type === "success"
                        ? "border-green-500"
                        : notification.type === "error"
                            ? "border-red-500"
                            : notification.type === "warning"
                                ? "border-yellow-500"
                                : "border-blue-500"
                    }`}
            >
                <div className="mr-3 text-2xl">
                    {notification.type === "success" && <i className="fas fa-check-circle text-green-500"></i>}
                    {notification.type === "error" && <i className="fas fa-exclamation-circle text-red-500"></i>}
                    {notification.type === "warning" && <i className="fas fa-exclamation-triangle text-yellow-500"></i>}
                    {notification.type === "info" && <i className="fas fa-info-circle text-blue-500"></i>}
                </div>
                <div>
                    <p className="text-white">{notification.message}</p>
                </div>
            </div>
        </div>
    );
}