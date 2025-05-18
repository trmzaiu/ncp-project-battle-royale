import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useWebSocketContext } from "../context/WebSocketContext";

export default function Auth() {
    const navigate = useNavigate();
    const { sendMessage, subscribe, isConnected } = useWebSocketContext();
    const [animationComplete, setAnimationComplete] = useState(false);
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

    useEffect(() => {
        const timer1 = setTimeout(() => setAnimationComplete(true), 600);
        const unsubscribe = subscribe((res) => {
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
        });

        return () => { unsubscribe(); clearTimeout(timer1); }
    }, [subscribe, navigate]);

    const handleLogin = () => {
        if (!username || !password)
            return showNotification("Please fill in both fields.", "warning");
        if (!isConnected)
            return showNotification("Not connected to server.", "error");

        sendMessage({
            type: "login",
            data: { username, password }
        });
    };

    const handleRegister = () => {
        if (!newUsername || !newPassword)
            return showNotification("Please fill in all fields.", "warning");
        if (!isConnected)
            return showNotification("Not connected to server.", "error");

        sendMessage({
            type: "register",
            data: {
                username: newUsername,
                password: newPassword
            }
        });
    };

    return (
        <div className="flex flex-col items-center justify-center min-h-screen w-full bg-gradient-to-b from-blue-900 via-blue-800 to-blue-900 overflow-hidden relative" style={{ fontFamily: "'ClashDisplay', sans-serif" }}>
            {/* Animated Background Elements */}
            <div className="absolute inset-0 overflow-hidden">
                <div className="absolute top-10 left-10 w-24 h-24 bg-pink-400 rounded-full opacity-50 animate-pulse"></div>
                <div className="absolute top-32 right-20 w-20 h-20 bg-red-400 rounded-full opacity-40 animate-bounce"></div>
                <div className="absolute bottom-20 left-32 w-28 h-28 bg-cyan-300 rounded-full opacity-60 animate-pulse"></div>
                <div className="absolute bottom-40 right-40 w-16 h-16 bg-lime-400 rounded-full opacity-50 animate-bounce"></div>
                <div className="absolute top-1/2 left-1/4 w-32 h-32 bg-yellow-300 rounded-full opacity-40 animate-pulse"></div>
                <div className="absolute bottom-1/3 right-1/4 w-24 h-24 bg-orange-400 rounded-full opacity-50 animate-bounce"></div>
            </div>

            <div className="relative z-10 w-full max-w-md">
                {/* Crown Logo */}
                {/* Crown animation that comes from top - responsive sizing */}
                <div className="relative flex justify-center mb-2 sm:mb-4">
                    <div
                        className={`transition-all duration-1000 ease-out transform
                        ${animationComplete ? 'translate-y-0 opacity-100' : '-translate-y-24 opacity-0'}`}
                    >
                        <img
                            className="w-45 pointer-events-none"
                            src="/assets/icon_crown.png"
                            alt=""
                        />
                    </div>

                    {/* Flash effect when crown lands - responsive sizing */}
                    <div
                        className={`absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 transition-all duration-500 delay-800
                        ${animationComplete ? 'opacity-100 scale-100' : 'opacity-0 scale-0'}`}
                    >
                        <div className="animate-ping opacity-70 duration-300">
                            <svg
                                className="w-32 h-32 sm:w-40 sm:h-40 md:w-52 md:h-52"
                                viewBox="0 0 200 200"
                                fill="none"
                                xmlns="http://www.w3.org/2000/svg"
                            >
                                <circle cx="100" cy="100" r="50" fill="white" fillOpacity="0.3" />
                                <circle cx="100" cy="100" r="30" fill="white" fillOpacity="0.5" />
                                {/* Light rays */}
                                <path d="M100 10L110 90L100 100L90 90L100 10Z" fill="white" fillOpacity="0.3" />
                                <path d="M100 190L110 110L100 100L90 110L100 190Z" fill="white" fillOpacity="0.3" />
                                <path d="M10 100L90 110L100 100L90 90L10 100Z" fill="white" fillOpacity="0.3" />
                                <path d="M190 100L110 110L100 100L110 90L190 100Z" fill="white" fillOpacity="0.3" />
                                <path d="M30 30L92 92L100 100L92 92L30 30Z" fill="white" fillOpacity="0.3" />
                                <path d="M170 170L108 108L100 100L108 108L170 170Z" fill="white" fillOpacity="0.3" />
                                <path d="M30 170L92 108L100 100L92 108L30 170Z" fill="white" fillOpacity="0.3" />
                                <path d="M170 30L108 92L100 100L108 92L170 30Z" fill="white" fillOpacity="0.3" />
                            </svg>
                        </div>
                    </div>
                </div>

                {/* Card Container */}
                <div className="bg-gradient-to-b from-blue-600/95 to-purple-600/95 rounded-xl shadow-2xl overflow-hidden border-4 border-yellow-400 transform transition-all duration-300">
                    {/* Header */}
                    <div className="flex flex-col items-center p-6 bg-gradient-to-b from-blue-800 to-blue-900 border-b-4 border-yellow-400">
                        <h1 className="text-4xl sm:text-5xl md:text-6xl font-extrabold text-transparent bg-clip-text bg-gradient-to-b from-blue-200 via-blue-100 to-yellow-200 mb-2 sm:mb-3 tracking-wide select-none drop-shadow-lg">
                            ROYAKA
                        </h1>
                        <div className="flex w-full rounded-lg overflow-hidden border-2 border-yellow-300">
                            <button
                                onClick={() => setActiveTab("login")}
                                className={`flex-1 py-4 text-center font-bold relative transition-all duration-300 ${activeTab === "login"
                                    ? "text-yellow-200 bg-gradient-to-b from-red-500 to-red-600"
                                    : "text-blue-100 hover:text-yellow-200 bg-gradient-to-b from-blue-700 to-blue-800"
                                    }`}
                            >
                                LOGIN
                            </button>
                            <button
                                onClick={() => setActiveTab("register")}
                                className={`flex-1 py-4 text-center font-bold relative transition-all duration-300 ${activeTab === "register"
                                    ? "text-yellow-200 bg-gradient-to-b from-red-500 to-red-600"
                                    : "text-blue-100 hover:text-yellow-200 bg-gradient-to-b from-blue-700 to-blue-800"
                                    }`}
                            >
                                REGISTER
                            </button>
                        </div>
                    </div>

                    {activeTab === "login" ? (
                        <div className="p-6 animate-fadeIn bg-gradient-to-b from-blue-800 to-blue-900">
                            <div className="text-center mb-6">
                                <h2 className="text-3xl font-bold text-yellow-300 mb-2">Welcome Back, Warrior!</h2>
                                <p className="text-cyan-200 text-base">Enter your credentials to join the battle</p>
                            </div>

                            <div className="relative mb-5 group">
                                <div className="absolute left-4 top-1/2 -translate-y-1/2 text-cyan-300 text-xl">
                                    👤
                                </div>
                                <input
                                    type="text"
                                    value={username}
                                    onChange={(e) => setUsername(e.target.value)}
                                    placeholder="Username"
                                    className="w-full py-4 pl-12 pr-4 border-2 border-cyan-400 bg-blue-700/70 text-cyan-100 rounded-lg text-base transition-all focus:border-yellow-400 focus:shadow-md focus:shadow-yellow-400/50 outline-none placeholder-stone-50"
                                />
                            </div>

                            <div className="relative mb-8 group">
                                <div className="absolute left-4 top-1/2 -translate-y-1/2 text-blue-900 text-xl">
                                    🔒
                                </div>
                                <input
                                    type="password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    placeholder="Password"
                                    className="w-full py-4 pl-12 pr-4 border-2 border-cyan-400 bg-blue-700/70 text-cyan-100 rounded-lg text-base transition-all focus:border-yellow-400 focus:shadow-md focus:shadow-yellow-400/50 outline-none placeholder-stone-50"
                                />
                            </div>

                            <button
                                onClick={handleLogin}
                                className="w-full py-4 bg-gradient-to-r from-yellow-400 to-orange-400 text-gradient-to-r from-blue-800 to-blue-400 rounded-lg font-extrabold text-xl cursor-pointer transition-all flex justify-center items-center gap-2 shadow-xl shadow-yellow-500/50 hover:shadow-2xl hover:shadow-yellow-500/60 hover:-translate-y-1 active:translate-y-0 border-2 border-yellow-300"
                            >
                                {loading ? (
                                    <div className="w-6 h-6 border-4 border-blue-800/30 border-t-blue-800 rounded-full animate-spin"></div>
                                ) : (
                                    <>
                                        <span className="font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-orange-50 to-yellow-50 tracking-wide select-none drop-shadow-lg">BATTLE NOW</span>
                                        <span className="text-2xl">⚔️</span>
                                    </>
                                )}
                            </button>
                        </div>
                    ) : (
                        <div className="p-6 animate-fadeIn bg-gradient-to-b from-blue-800 to-blue-900">
                            <div className="text-center mb-6">
                                <h2 className="text-3xl font-bold text-yellow-300 mb-2">Join The Arena!</h2>
                                <p className="text-cyan-200 text-base">Create your warrior account</p>
                            </div>

                            <div className="relative mb-5 group">
                                <div className="absolute left-4 top-1/2 -translate-y-1/2 text-cyan-300 text-xl">
                                    👤
                                </div>
                                <input
                                    type="text"
                                    value={newUsername}
                                    onChange={(e) => setNewUsername(e.target.value)}
                                    placeholder="Choose a username"
                                    className="w-full py-4 pl-12 pr-4 border-2 border-cyan-400 bg-blue-700/70 text-cyan-100 rounded-lg text-base transition-all focus:border-yellow-400 focus:shadow-md focus:shadow-yellow-400/50 outline-none placeholder-stone-50"
                                />
                            </div>

                            <div className="relative mb-8 group">
                                <div className="absolute left-4 top-1/2 -translate-y-1/2 text-cyan-300 text-xl">
                                    🔒
                                </div>
                                <input
                                    type="password"
                                    value={newPassword}
                                    onChange={(e) => setNewPassword(e.target.value)}
                                    placeholder="Create password"
                                    className="w-full py-4 pl-12 pr-4 border-2 border-cyan-400 bg-blue-700/70 text-cyan-100 rounded-lg text-base transition-all focus:border-yellow-400 focus:shadow-md focus:shadow-yellow-400/50 outline-none placeholder-stone-50"
                                />
                            </div>

                            <button
                                onClick={handleRegister}
                                className="w-full py-4 bg-gradient-to-r from-yellow-400 to-orange-400 text-blue-900 rounded-lg font-extrabold text-xl cursor-pointer transition-all flex justify-center items-center gap-2 shadow-xl shadow-yellow-500/50 hover:shadow-2xl hover:shadow-yellow-500/60 hover:-translate-y-1 active:translate-y-0 border-2 border-yellow-300"
                            >
                                {loading ? (
                                    <div className="w-6 h-6 border-4 border-blue-800/30 border-t-blue-800 rounded-full animate-spin"></div>
                                ) : (
                                    <>
                                        <span className="font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-orange-50 to-yellow-50 tracking-wide select-none drop-shadow-lg">CREATE WARRIOR</span>
                                        <span className="text-2xl">🛡️</span>
                                    </>
                                )}
                            </button>
                        </div>
                    )}
                </div>

                {/* Decorative Elements */}
                <div className="flex justify-center mt-8">
                    <div className="w-20 h-20 bg-gradient-to-br from-cyan-500 to-blue-500 rounded-full border-4 border-yellow-300 shadow-lg shadow-cyan-500/50 flex items-center justify-center mx-3 animate-bounce">
                        <span className="text-3xl">🏆</span>
                    </div>
                    <div className="w-20 h-20 bg-gradient-to-br from-fuchsia-500 to-purple-500 rounded-full border-4 border-yellow-300 shadow-lg shadow-fuchsia-500/50 flex items-center justify-center mx-3 animate-pulse">
                        <span className="text-3xl">⚔️</span>
                    </div>
                    <div className="w-20 h-20 bg-gradient-to-br from-orange-500 to-red-500 rounded-full border-4 border-yellow-300 shadow-lg shadow-orange-500/50 flex items-center justify-center mx-3 animate-bounce">
                        <span className="text-3xl">🛡️</span>
                    </div>
                </div>
            </div>

            {/* Notification */}
            <div
                className={`fixed top-5 right-5 flex items-center justify-center bg-blue-700/90 backdrop-blur-md rounded-lg py-4 px-5 text-white text-base shadow-xl border-l-4 z-50 transform transition-transform duration-300 ${notification.show ? "translate-x-0" : "translate-x-full"
                    } ${notification.type === "success"
                        ? "border-yellow-300"
                        : notification.type === "error"
                            ? "border-red-400"
                            : notification.type === "warning"
                                ? "border-orange-400"
                                : "border-blue-400"
                    }`}
            >
                <div className="mr-4 text-3xl">
                    {notification.type === "success" && <span className="text-yellow-300">✓</span>}
                    {notification.type === "error" && <span className="text-red-400">✕</span>}
                    {notification.type === "warning" && <span className="text-orange-400">⚠</span>}
                    {notification.type === "info" && <span className="text-blue-400">ℹ</span>}
                </div>
                <div>
                    <p className="text-white font-bold">{notification.message}</p>
                </div>
            </div>
        </div>
    );
}