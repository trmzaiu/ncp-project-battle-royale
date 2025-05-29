import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useWebSocketContext } from "../context/WebSocketContext";

export default function Lobby() {
    const url = process.env.NODE_ENV === 'production' ? "/royaka-2025-fe/" : "/";
    const navigate = useNavigate();
    const { sendMessage, subscribe } = useWebSocketContext();

    const [user, setUser] = useState({});
    const [expMax, setExpMax] = useState(100);
    const [selectedMode, setSelectedMode] = useState("simple");
    const [notification, setNotification] = useState({ show: false, message: "" });
    const [isJoiningGame, setIsJoiningGame] = useState(false);
    const selectedModeRef = useRef(selectedMode);

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
            selectedModeRef.current = selectedMode;
            switch (res.type) {
                case "user_response":
                    if (res.success) {
                        setUser(res.data.user);
                        setExpMax(res.data.maxExp);
                        localStorage.setItem("username", res.data.user.username);
                    } else {

                        showNotification("Failed to fetch user data");
                    }
                    break;

                case "match_found":
                    if (res.success) {
                        localStorage.setItem("room_id", res.data.room_id);
                        showNotification("Match found! Starting game...");
                        if (selectedModeRef.current === "simple") {
                            setTimeout(() => navigate("/game-simple"), 1000);
                        } else {
                            setTimeout(() => navigate("/game-enhanced"), 1000);
                        }
                    } else {
                        showNotification("Failed to find a match. Please try again.");
                        setIsJoiningGame(false);
                    }
                    break;

                case "match_timeout":
                    showNotification("Match timed out. Waiting for new match...");
                    setIsJoiningGame(false);
                    break;

                default:
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
    }, [subscribe, sendMessage, navigate, selectedMode]);

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

        showNotification("Finding a match...");
    }

    function handleLogout() {
        localStorage.removeItem("session_id");
        navigate("/auth");
    }

    function viewLeaderboard() {
        navigate("/card-desk");
    }

    const expProgressStyle = {
        width: `${(user.exp / expMax) * 100}%`,
    };

    return (
        <div
            className="min-h-screen bg-gradient-to-br from-sky-500 to-blue-600 p-4 md:p-8 font-sans"
            style={{ fontFamily: "'ClashDisplay', sans-serif" }}
        >
            {/* Gold and Gems Bar */}
            <div className="flex justify-end mb-4 space-x-2">
                <div className="bg-blue-900 px-3 py-1 rounded-full border-2 border-yellow-400 flex items-center shadow-lg">
                    <span className="text-yellow-400 mr-1 text-xl">💰</span>
                    <span className="font-bold text-yellow-300 mt-1">{(user.gold ?? 0).toLocaleString('en-US')}</span>
                </div>
                <div className="bg-blue-900 px-3 py-1 rounded-full border-2 border-yellow-400 flex items-center shadow-lg">
                    <span className="text-green-400 mr-1 text-xl">💎</span>
                    <span className="font-bold text-green-300 mt-1">436</span>
                </div>
            </div>

            <div className="max-w-6xl mx-auto">
                {/* Header */}
                <div className="text-center mb-8">
                    <div className="relative inline-block">
                        <h1
                            className="text-5xl md:text-6xl font-black text-white mb-2 drop-shadow-lg pointer-events-none"
                            style={{ textShadow: "3px 3px 0 #2563eb, 6px 6px 0 #1d4ed8" }}
                        >
                            <span className="text-yellow-400">ROY</span>
                            <span className="text-red-500">AKA</span>
                        </h1>
                        <img
                            className="absolute w-10 -top-4 -right-4 transform rotate-20 pointer-events-none drop-shadow-[0_0_10px_rgba(255,255,0,0.6)]"
                            src={`${url}assets/icon_crown.png`}
                            alt=""
                        />
                        <img
                            className="absolute w-12 -bottom-2 -left-6 transform -rotate-12 pointer-events-none drop-shadow-[0_0_10px_rgba(255,255,0,0.6)]"
                            src={`${url}assets/icon_badge.png`}
                            alt=""
                        />
                    </div>
                    <div className="bg-blue-900 inline-block px-6 py-2 rounded-xl border-4 border-yellow-400 shadow-lg transform -rotate-1 pointer-events-none">
                        <p className="text-lg text-yellow-300">
                            Epic Tower Battles Await!
                        </p>
                    </div>
                </div>

                {/* Main Container */}
                <div className="flex flex-col md:flex-row gap-6">
                    {/* Player Card */}
                    <div className="bg-gradient-to-b from-blue-800 to-blue-900 rounded-xl shadow-lg p-6 flex-1 border-4 border-yellow-400 relative overflow-hidden">
                        <img
                            className="absolute w-24 -right-4 -bottom-4 opacity-60 pointer-events-none"
                            src={`${url}assets/icon_decorate.png`}
                            alt=""
                        />
                        <div className="flex flex-col items-center">
                            <div className="relative w-35 h-35 mb-4">
                                <div className="w-40 rounded-full bg-gradient-to-br from-cyan-400 to-blue-500 border-4 border-yellow-400 shadow-lg overflow-hidden relative">
                                    <img
                                        src={`${url}assets/avatars/Avatar${user.avatar}.png`}
                                        alt="avatar"
                                        className="w-full h-full object-cover"
                                    />
                                </div>

                                {/* Level Badge - overlaps bottom-right */}
                                <div className="absolute -bottom-2 right-4 w-10 h-10 z-10 shadow-md drop-shadow-[0_0_10px_rgba(255,255,0,0.4)]">
                                    <div className="relative w-full h-full flex items-center justify-center">
                                        <img
                                            className="pointer-events-none w-full h-full"
                                            src={`${url}assets/icon_banner.png`}
                                            alt=""
                                        />
                                        <span className="absolute text-white text-base font-bold">
                                            {user.level}
                                        </span>
                                    </div>
                                </div>

                            </div>

                            <div className="w-full bg-gradient-to-r from-violet-500 to-blue-500 py-2 px-4 rounded-xl border-4 border-cyan-300 mb-4 transform rotate-1 shadow-lg">
                                <h2 className="text-2xl font-black text-white text-center">
                                    {user.username}
                                </h2>
                            </div>

                            <div className="w-full space-y-4">
                                {/* Trophy Count */}
                                <div className="bg-gradient-to-r from-yellow-500 to-amber-600 p-3 rounded-xl border-4 border-yellow-300 shadow-md flex items-center justify-between">
                                    <div className="flex items-center">
                                        <span className="text-3xl mr-2">🏆</span>
                                        <span className="text-white">Trophies</span>
                                    </div>
                                    <span className="font-black text-white text-2xl">
                                        {user.gamesWon}
                                    </span>
                                </div>

                                {/* Experience Bar */}
                                <div className="bg-blue-700 rounded-xl p-3 border-4 border-cyan-400 shadow-md">
                                    <div className="flex justify-between mb-2">
                                        <span className="text-cyan-300">
                                            Experience
                                        </span>
                                        <span className="font-bold text-white">
                                            {user.exp}/{expMax}
                                        </span>
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
                                        <span className="text-green-300 text-sm">
                                            Games Played
                                        </span>
                                        <div className="text-white text-2xl">
                                            {user.gamesPlayed}
                                        </div>
                                    </div>
                                    <div className="bg-gradient-to-b from-purple-600 to-purple-700 rounded-xl p-3 text-center border-4 border-purple-400 shadow-md transform rotate-1">
                                        <span className="text-purple-300 text-sm">
                                            Victories
                                        </span>
                                        <div className="text-white text-2xl">
                                            {user.gamesWon}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Game Options */}
                    <div className="bg-gradient-to-b from-blue-800 to-blue-900 rounded-xl shadow-lg p-6 flex-1 border-4 border-yellow-400 relative overflow-hidden">
                        <img
                            className="absolute w-36 -left-8 -top-8 opacity-40 -rotate-12 pointer-events-none"
                            src={`${url}assets/icon_timed_match.png`}
                            alt=""
                        />

                        <div className="bg-gradient-to-r from-red-500 to-orange-500 py-2 px-4 rounded-xl border-4 border-yellow-300 mb-6 shadow-lg transform -rotate-1  pointer-events-none">
                            <h3 className="text-2xl font-black text-center text-white">
                                BATTLE MODES
                            </h3>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
                            {/* TURN-BASED Mode Card */}
                            <div
                                className={`${selectedMode === "simple"
                                    ? "bg-gradient-to-b from-red-500 to-red-600 border-yellow-300 shadow-lg transform scale-105"
                                    : "bg-gradient-to-b from-blue-700 to-blue-800 border-blue-400"
                                    } rounded-xl p-4 cursor-pointer transition-all hover:shadow-lg border-4 relative overflow-hidden flex flex-col items-center`}
                                onClick={() => setSelectedMode("simple")}
                            >
                                <img
                                    className=" w-20 mb-2 pointer-events-none"
                                    src={`${url}assets/icon_turn_based.png`}
                                    alt=""
                                />
                                <div className="text-center mb-1 text-white text-xl">
                                    TURN-BASED
                                </div>
                                <div className="text-xs text-center text-white">
                                    Strategic 1v1, move by move.
                                </div>
                            </div>

                            {/* TIMED MATCH Mode Card */}
                            <div
                                className={`${selectedMode === "enhanced"
                                    ? "bg-gradient-to-b from-red-500 to-red-600 border-yellow-300 shadow-lg transform scale-105"
                                    : "bg-gradient-to-b from-blue-700 to-blue-800 border-blue-400"
                                    } rounded-xl p-4 cursor-pointer transition-all hover:shadow-lg border-4 relative overflow-hidden flex flex-col items-center`}
                                onClick={() => setSelectedMode("enhanced")}
                            >
                                <img
                                    className=" w-20 mb-2 pointer-events-none"
                                    src={`${url}assets/icon_timed_match.png`}
                                    alt=""
                                />
                                <div className="text-center mb-1 text-white text-xl">
                                    TIMED MATCH
                                </div>
                                <div className="text-xs text-center text-white">
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
                                        <svg
                                            className="w-6 h-6 animate-spin text-white"
                                            xmlns="http://www.w3.org/2000/svg"
                                            fill="none"
                                            viewBox="0 0 24 24"
                                        >
                                            <circle
                                                className="opacity-25"
                                                cx="12"
                                                cy="12"
                                                r="10"
                                                stroke="currentColor"
                                                strokeWidth="4"
                                            ></circle>
                                            <path
                                                className="opacity-75"
                                                fill="currentColor"
                                                d="M4 12a8 8 0 018-8v8z"
                                            ></path>
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
                                    TROOPS DECK
                                </span>
                            </button>

                            {/* Logout Button */}
                            <button
                                onClick={handleLogout}
                                className="w-full py-3 bg-gradient-to-r from-violet-600 to-violet-500 text-white rounded-xl font-black text-xl transition-all hover:shadow-lg hover:-translate-y-1 active:translate-y-0 border-4 border-pink-300 shadow-md"
                            >
                                <span className="flex items-center justify-center">
                                    <span className="text-xl mr-2">🚪</span>
                                    EXIT ARENA
                                </span>
                            </button>
                        </div>
                    </div>
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
                <img
                    className="fixed w-28 top-4 left-4 animate-pulse pointer-events-none"
                    src={`${url}assets/icon_badge.png`}
                    alt=""
                />
                <img
                    className="fixed w-20 bottom-4 right-4 animate-bounce pointer-events-none"
                    src={`${url}assets/icon_timed_match.png`}
                    alt=""
                />
            </div>
        </div>
    );
}