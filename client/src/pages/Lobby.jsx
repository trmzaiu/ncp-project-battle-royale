import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useWebSocketContext } from "../context/WebSocketContext";

export default function Lobby() {
    const navigate = useNavigate();
    const { sendMessage, subscribe } = useWebSocketContext();

    const [user, setUser] = useState({});
    const [expMax, setExpMax] = useState(100);
    const [selectedMode, setSelectedMode] = useState("simple");
    const [notification, setNotification] = useState({ show: false, message: "" });
    const [logs, setLogs] = useState([
        { timestamp: getCurrentTimestamp(), type: "SYSTEM", message: "Welcome to Royaka!" },
    ]);
    const [isJoiningGame, setIsJoiningGame] = useState(false);

    const showNotification = (message) => {
        setNotification({ show: true, message });
        setTimeout(() => setNotification((prev) => ({ ...prev, show: false })), 4000);
    };

    useEffect(() => {
        if (!localStorage.getItem("session_id")) {
            showNotification("Session expired. Redirecting to login...");
            setTimeout(() => navigate("/auth"), 1500);
            return;
        }

        const unsubscribe = subscribe((res) => {
            switch (res.type) {
                case "user_response":
                    if (res.success) {
                        setUser(res.data.user);
                        setExpMax(res.data.maxExp);
                        localStorage.setItem("username", res.data.user.username);
                        addLog("SYSTEM", "Fetched user data successfully!");
                    } else {
                        addLog("ERROR", res.message || "Failed to fetch user data");
                        showNotification(res.message || "Failed to fetch user data");
                    }
                    break;

                case "match_found":
                    if (res.success) {
                        localStorage.setItem("room_id", res.data.room_id);
                        showNotification("Match found! Starting game...");
                        addLog("GAME", "Match found! Starting game...");
                        setTimeout(() => navigate("/game"), 1000);
                    } else {
                        addLog("ERROR", res.message || "Failed to find a match");
                        showNotification(res.message || "Failed to find a match. Please try again.");
                        setIsJoiningGame(false);
                    }
                    break;

                case "match_timeout":
                    addLog("SYSTEM", "Match timed out. Waiting for new match...");
                    showNotification("Match timed out. Waiting for new match...");
                    setIsJoiningGame(false);
                    break;

                default:
                    if (res.message) showNotification(res.message);
                    break;
            }
        });

        const sessionId = localStorage.getItem("session_id");
        if (sessionId) {
            sendMessage({
                type: "get_user",
                data: { session_id: sessionId },
            });
        }

        return () => unsubscribe();
    }, [subscribe, sendMessage, navigate]);

    // === UI action handlers ===
    function findMatch() {
        if (isJoiningGame) {
            showNotification("Already searching for a match...");
            return;
        }

        if (!user) {
            showNotification("User data not loaded yet. Please wait...");
            return;
        }

        setIsJoiningGame(true);
        sendMessage({
            type: "find_match",
            data: { username: user.username, mode: selectedMode },
        });

        addLog("GAME", `Finding ${selectedMode} mode match...`);
        showNotification("Finding a match...");
    }

    function handleLogout() {
        localStorage.removeItem("session_id");
        addLog("SYSTEM", "Logging out...");
        navigate("/auth");
    }

    function viewLeaderboard() {
        addLog("SYSTEM", "Opening leaderboard...");
    }

    const expProgressStyle = {
        width: `${(user.exp / expMax) * 100}%`,
    };

    // === Helper functions ===
    function getCurrentTimestamp() {
        const now = new Date();
        return now.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    }

    function addLog(type, message) {
        setLogs((prev) => [
            ...prev,
            { timestamp: getCurrentTimestamp(), type, message },
        ]);
    }

    return (
        <div className="min-h-screen bg-gradient-to-br from-sky-500 to-blue-600 p-4 md:p-8 font-sans" style={{ fontFamily: "'ClashDisplay', sans-serif" }}>
            {/* Gold and Gems Bar */}
            <div className="flex justify-end mb-4 space-x-2">
                <div className="bg-blue-900 px-3 py-1 rounded-full border-2 border-yellow-400 flex items-center shadow-lg">
                    <span className="text-yellow-400 mr-1 text-xl">💰</span>
                    <span className="font-bold text-yellow-300">12,850</span>
                </div>
                <div className="bg-blue-900 px-3 py-1 rounded-full border-2 border-yellow-400 flex items-center shadow-lg">
                    <span className="text-green-400 mr-1 text-xl">💎</span>
                    <span className="font-bold text-green-300">436</span>
                </div>
            </div>

            <div className="max-w-6xl mx-auto">
                {/* Header */}
                <div className="text-center mb-8">
                    <div className="relative inline-block">
                        <h1 className="text-5xl md:text-6xl font-black text-white mb-2 drop-shadow-lg" style={{ textShadow: "3px 3px 0 #2563eb, 6px 6px 0 #1d4ed8" }}>
                            <span className="text-yellow-400">ROY</span>
                            <span className="text-red-500">AKA</span>
                        </h1>
                        <div className="absolute -top-4 -right-4 transform rotate-12 text-2xl">✨</div>
                        <div className="absolute -bottom-2 -left-4 transform -rotate-12 text-2xl">✨</div>
                    </div>
                    <div className="bg-blue-900 inline-block px-6 py-2 rounded-xl border-4 border-yellow-400 shadow-lg transform -rotate-1">
                        <p className="text-lg text-yellow-300 font-bold">
                            Epic Tower Battles Await!
                        </p>
                    </div>
                </div>

                {/* Main Container */}
                <div className="flex flex-col md:flex-row gap-6">
                    {/* Player Card */}
                    <div className="bg-gradient-to-b from-blue-800 to-blue-900 rounded-xl shadow-lg p-6 flex-1 border-4 border-yellow-400 relative overflow-hidden">
                        <div className="absolute -right-6 -bottom-6 text-yellow-400/20 text-8xl">👑</div>
                        <div className="flex flex-col items-center">
                            <div className="relative w-28 h-28 mb-4">
                                <div className="w-full h-full rounded-full bg-gradient-to-br from-cyan-400 to-blue-500 border-4 border-yellow-400 shadow-lg overflow-hidden relative">
                                    <img
                                        src={user.avatar}
                                        alt="avatar"
                                        className="w-full h-full object-cover"
                                    />
                                </div>

                                {/* Level Badge - overlaps bottom-right */}
                                <div className="absolute -bottom-1 -right-1 bg-red-500 w-9 h-9 rounded-full border-2 border-white flex items-center justify-center z-10 shadow-md">
                                    <span className="text-white font-bold text-base">{user.level}</span>
                                </div>
                            </div>

                            <div className="w-full bg-gradient-to-r from-purple-500 to-blue-500 py-2 px-4 rounded-xl border-4 border-cyan-300 mb-4 transform rotate-1 shadow-lg">
                                <h2 className="text-2xl font-black text-white text-center">{user.username}</h2>
                            </div>

                            <div className="w-full space-y-4">
                                {/* Trophy Count */}
                                <div className="bg-gradient-to-r from-yellow-500 to-amber-600 p-3 rounded-xl border-4 border-yellow-300 shadow-md flex items-center justify-between">
                                    <div className="flex items-center">
                                        <span className="text-3xl mr-2">🏆</span>
                                        <span className="text-white font-bold">Trophies</span>
                                    </div>
                                    <span className="font-black text-white text-2xl">{user.gamesWon}</span>
                                </div>

                                {/* Experience Bar */}
                                <div className="bg-blue-700 rounded-xl p-3 border-4 border-cyan-400 shadow-md">
                                    <div className="flex justify-between mb-2">
                                        <span className="text-cyan-300 font-bold">Experience</span>
                                        <span className="font-bold text-white">{user.exp}/{expMax}</span>
                                    </div>

                                    <div className="w-full h-4 bg-blue-900 rounded-full overflow-hidden border-2 border-cyan-300">
                                        <div
                                            className="h-full bg-gradient-to-r from-cyan-400 to-blue-400 transition-all duration-500"
                                            style={expProgressStyle}
                                        />
                                    </div>
                                </div>

                                {/* Stats */}
                                <div className="grid grid-cols-2 gap-3">
                                    <div className="bg-gradient-to-b from-green-600 to-green-700 rounded-xl p-3 text-center border-4 border-green-400 shadow-md transform -rotate-1">
                                        <span className="text-green-300 text-sm font-bold">Games Played</span>
                                        <div className="text-white font-black text-2xl">{user.gamesPlayed}</div>
                                    </div>
                                    <div className="bg-gradient-to-b from-purple-600 to-purple-700 rounded-xl p-3 text-center border-4 border-purple-400 shadow-md transform rotate-1">
                                        <span className="text-purple-300 text-sm font-bold">Victories</span>
                                        <div className="text-white font-black text-2xl">{user.gamesWon}</div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Game Options */}
                    <div className="bg-gradient-to-b from-blue-800 to-blue-900 rounded-xl shadow-lg p-6 flex-1 border-4 border-yellow-400 relative overflow-hidden">
                        <div className="absolute -left-8 -top-8 text-red-400/20 text-8xl">⚔️</div>

                        <div className="bg-gradient-to-r from-red-500 to-orange-500 py-2 px-4 rounded-xl border-4 border-yellow-300 mb-6 shadow-lg transform -rotate-1">
                            <h3 className="text-2xl font-black text-center text-white">BATTLE MODES</h3>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
                            {/* TURN-BASED Mode Card */}
                            <div
                                className={`${selectedMode === "simple"
                                    ? "bg-gradient-to-b from-red-500 to-red-600 border-yellow-300 shadow-lg transform scale-105"
                                    : "bg-gradient-to-b from-blue-700 to-blue-800 border-blue-400"
                                    } rounded-xl p-4 cursor-pointer transition-all hover:shadow-lg border-4 relative overflow-hidden`}
                                onClick={() => setSelectedMode("simple")}
                            >
                                <div className="absolute -right-4 -bottom-4 text-white/10 text-6xl">🏆</div>
                                <div className="text-5xl text-center mb-2">🏰</div>
                                <div className="font-black text-center mb-1 text-white text-xl">TURN-BASED</div>
                                <div className="text-sm text-center text-white font-semibold">
                                    Strategic 1v1, move by move.
                                </div>
                            </div>

                            {/* TIMED MATCH Mode Card */}
                            <div
                                className={`${selectedMode === "enhanced"
                                    ? "bg-gradient-to-b from-red-500 to-red-600 border-yellow-300 shadow-lg transform scale-105"
                                    : "bg-gradient-to-b from-blue-700 to-blue-800 border-blue-400"
                                    } rounded-xl p-4 cursor-pointer transition-all hover:shadow-lg border-4 relative overflow-hidden`}
                                onClick={() => setSelectedMode("enhanced")}
                            >
                                <div className="absolute -right-4 -bottom-4 text-white/10 text-6xl">👥</div>
                                <div className="text-5xl text-center mb-2">⚔️</div>
                                <div className="font-black text-center mb-1 text-white text-xl">TIMED MATCH</div>
                                <div className="text-sm text-center text-white font-semibold">
                                    Fast-paced 1v1 chaos.
                                </div>
                            </div>
                        </div>

                        <div className="space-y-4">
                            {/* Battle Button */}
                            <button
                                onClick={findMatch}
                                disabled={isJoiningGame}
                                className={`w-full py-4 rounded-xl font-black text-2xl transition-all shadow-lg
                                ${isJoiningGame
                                        ? "bg-gray-600 cursor-not-allowed"
                                        : "bg-gradient-to-r from-red-500 to-red-600 text-white hover:-translate-y-1 active:translate-y-0 border-4 border-yellow-300"
                                    }`}
                            >
                                {isJoiningGame ? (
                                    <span className="flex items-center justify-center gap-2">
                                        <svg className="w-6 h-6 animate-spin text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z"></path>
                                        </svg>
                                        MATCHING...
                                    </span>
                                ) : (
                                    <span className="flex items-center justify-center">
                                        <span className="text-2xl mr-2">⚔️</span>
                                        BATTLE!
                                    </span>
                                )}
                            </button>

                            {/* Leaderboard Button */}
                            <button
                                onClick={viewLeaderboard}
                                className="w-full py-3 bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-xl font-black text-xl transition-all hover:shadow-lg hover:-translate-y-1 active:translate-y-0 border-4 border-cyan-300 shadow-md"
                            >
                                <span className="flex items-center justify-center">
                                    <span className="text-xl mr-2">🏆</span>
                                    LEADERBOARD
                                </span>
                            </button>

                            {/* Logout Button */}
                            <button
                                onClick={handleLogout}
                                className="w-full py-3 bg-gradient-to-r from-purple-500 to-purple-600 text-white rounded-xl font-black text-xl transition-all hover:shadow-lg hover:-translate-y-1 active:translate-y-0 border-4 border-pink-300 shadow-md"
                            >
                                <span className="flex items-center justify-center">
                                    <span className="text-xl mr-2">🚪</span>
                                    EXIT ARENA
                                </span>
                            </button>
                        </div>
                    </div>
                </div>

                {/* Log Container */}
                <div className="mt-6 bg-blue-900 rounded-xl p-4 border-4 border-cyan-400 shadow-lg h-40 overflow-y-auto">
                    <div className="text-center mb-2">
                        <span className="bg-blue-700 px-4 py-1 rounded-full text-white text-sm font-black border-2 border-cyan-300">BATTLE LOG</span>
                    </div>
                    {logs.map((log, index) => {
                        let bgColor;
                        let borderColor;
                        let icon;

                        if (log.type === "SYSTEM") {
                            bgColor = "bg-gradient-to-r from-blue-600 to-blue-700";
                            borderColor = "border-blue-400";
                            icon = "🔧";
                        } else if (log.type === "GAME") {
                            bgColor = "bg-gradient-to-r from-green-600 to-green-700";
                            borderColor = "border-green-400";
                            icon = "⚔️";
                        } else {
                            bgColor = "bg-gradient-to-r from-yellow-600 to-yellow-700";
                            borderColor = "border-yellow-400";
                            icon = "📣";
                        }

                        return (
                            <div key={index} className={`mb-2 text-sm ${bgColor} p-2 rounded-lg border-2 ${borderColor} transform ${index % 2 === 0 ? 'rotate-1' : '-rotate-1'}`}>
                                <span className="text-white mr-2 font-bold">{log.timestamp}</span>
                                <span className="font-black mr-2 text-white">
                                    {icon} {log.type}
                                </span>
                                <span className="text-white font-semibold">{log.message}</span>
                            </div>
                        );
                    })}
                </div>

                {/* Notification */}
                {notification.show && (
                    <div className="fixed bottom-6 left-1/2 transform -translate-x-1/2 bg-gradient-to-r from-blue-600 to-purple-600 text-white px-6 py-3 rounded-xl shadow-lg border-4 border-yellow-300 animate-bounce">
                        <div className="flex items-center">
                            <span className="text-yellow-300 mr-2 text-2xl">👑</span>
                            <span className="font-black">{notification.message}</span>
                        </div>
                    </div>
                )}

                {/* Decorative elements */}
                <div className="fixed top-4 left-4 text-4xl animate-pulse">🏆</div>
                <div className="fixed bottom-4 right-4 text-4xl animate-bounce">⚔️</div>
            </div>
        </div>
    );
}