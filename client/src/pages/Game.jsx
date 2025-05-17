import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useWebSocketContext } from "../context/WebSocketContext";

export default function Game() {
    const navigate = useNavigate();
    const { sendMessage, subscribe } = useWebSocketContext();

    const [user, setUser] = useState({});
    const [opponent, setOpponent] = useState({});
    const [isGameInitialized, setIsGameInitialized] = useState(false);

    const [game, setGame] = useState(getInitialGameState());
    const [notification, setNotification] = useState({ show: false, message: "" });
    const [logs, setLogs] = useState([
        { timestamp: getCurrentTimestamp(), type: "SYSTEM", message: "Welcome to Royaka!" },
    ]);

    // === Initial Game State ===
    function getInitialGameState() {
        return {
            playerMana: 0,
            maxMana: 10,
            playerHealth: { king: 0, guard1: 0, guard2: 0 },
            opponentHealth: { king: 0, guard1: 0, guard2: 0 },
            playerShield: { king: 0, guard1: 0, guard2: 0 },
            opponentShield: { king: 0, guard1: 0, guard2: 0 },
            troops: {},
            selectedTroop: null,
            selectedTarget: null,
            playerTurn: null,
            gameOver: false,
        };
    }

    // === Notification Helper ===
    const showNotification = (message) => {
        setNotification({ show: true, message });
        setTimeout(() => setNotification((prev) => ({ ...prev, show: false })), 4000);
    };

    // === Effect: Initial Setup & WebSocket Subscription ===
    useEffect(() => {
        if (!localStorage.getItem("session_id")) {
            showNotification("Session expired. Redirecting to login...");
            setTimeout(() => navigate("/auth"), 1500);
            return;
        }

        const unsubscribe = subscribe(handleMessage);
        sendMessage({
            type: "get_game",
            data: {
                room_id: localStorage.getItem("room_id"),
                username: localStorage.getItem("username"),
            },
        });

        return () => unsubscribe();
    }, [subscribe, sendMessage, navigate]);

    // === Message Handler ===
    const handleMessage = (res) => {
        switch (res.type) {
            case "game_response":
                console.log("Game Response:", res);
                if (res.success) {
                    setUser(res.data.user);
                    setOpponent(res.data.opponent);

                    if (!isGameInitialized) {
                        initializeGame(res.data.turn, res.data.user.troops, res.data.user, res.data.opponent);
                    }
                } else {
                    showNotification(res.error || "Failed to get game data");
                }
                break;

            case "attack_response":
                console.log("Attack Response:", res);
                res.success ? handleAttackResponse(res) : showNotification(res.error || "Attack failed");
                break;

            case "game_over_response":
                res.success && handleGameOver();
                break;

            default:
                res.message && showNotification(res.message);
        }
    };

    // === Game Initialization ===
    const initializeGame = (turn, troops, userData, opponentData) => {
        setGame({
            playerMana: userData.mana,
            maxMana: 10,
            playerHealth: extractHP(userData.towers),
            opponentHealth: extractHP(opponentData.towers),
            playerShield: extractHP(userData.towers),
            opponentShield: extractHP(opponentData.towers),
            troops: troops.reduce((acc, troop) => {
                acc[troop.name] = troop;
                return acc;
            }, {}),
            selectedTroop: null,
            selectedTarget: null,
            playerTurn: turn,
            gameOver: false,
        });

        addLog("SYSTEM", "Game started.");
        addLog("SYSTEM", turn === userData?.user?.username ? "Your turn." : "Waiting for opponent's turn...");
        setIsGameInitialized(true);
    };

    const extractHP = (towers) => ({
        king: towers.king.max_hp,
        guard1: towers.guard1.max_hp,
        guard2: towers.guard2.max_hp,
    });

    // === Skip Turn ===
    const handleSkipTurn = () => {
        if (game.playerTurn !== user.user?.username) {
            return showNotification("Not your turn.");
        }

        sendMessage({
            type: "skip_turn",
            data: {
                room_id: localStorage.getItem("room_id"),
                username: localStorage.getItem("username"),
            },
        });

        addLog("ACTION", "Turn skipped.");
    };

    // === Select Troop ===
    const selectTroop = (troopName) => {
        if (game.playerTurn !== user.user?.username) return showNotification("Not your turn.");

        const troop = game.troops[troopName];
        if (game.playerMana < troop.mana) return showNotification("Not enough mana.");

        setGame((prev) => ({ ...prev, selectedTroop: troop }));
    };

    // === Select Target & Attack ===
    const selectTarget = (target) => {
        const { selectedTroop, playerMana, playerTurn } = game;
        const currentUser = user.user?.username;

        if (playerTurn !== currentUser || !selectedTroop) return showNotification("Invalid action.");
        if (playerMana < selectedTroop.mana) return showNotification("Not enough mana.");

        sendMessage({
            type: "attack",
            data: {
                troop: selectedTroop.name,
                target,
                room_id: localStorage.getItem("room_id"),
                username: currentUser,
            },
        });

        setGame((prev) => ({ ...prev, selectedTroop: null }));
    };

    // === Handle Attack Response ===
    const handleAttackResponse = (msg) => {
        const { attacker, defender, damage, target, isDestroyed, turn, troop } = msg.data;
        const isMe = attacker.user.username === localStorage.getItem("username");

        setGame((prev) => {
            const newState = { ...prev };

            if (isMe) {
                setUser(attacker);
                setOpponent(defender);
                newState.playerMana = attacker.mana;
                newState.opponentHealth[target] = defender.towers[target].hp;
                addLog("ACTION", `Your ${troop} dealt ${damage} damage to opponent's ${target}.`);
            } else {
                setUser(defender);
                setOpponent(attacker);
                newState.playerMana = defender.mana;
                newState.playerHealth[target] = defender.towers[target].hp;
                addLog("ACTION", `Opponent's ${troop} dealt ${damage} damage to your ${target}.`);
            }

            newState.playerTurn = turn;
            return newState;
        });

        if (isDestroyed) addLog("TOWER", `Tower ${target} has been destroyed!`);
        addLog("SYSTEM", turn === user?.user?.username ? "Your turn." : "Waiting for opponent's turn...");

        sendMessage({
            type: "game_over",
            data: {
                room_id: localStorage.getItem("room_id"),
            },
        });
    };

    // === Handle Game Over ===
    const handleGameOver = () => {
        setGame((prev) => ({ ...prev, gameOver: true }));
    };

    // === Play Again ===
    const handlePlayAgain = () => {
        setGame(getInitialGameState());
        setIsGameInitialized(false);

        sendMessage({
            type: "play_again",
            data: {
                session_id: localStorage.getItem("session_id"),
            },
        });
    };

    // === Logging Utility ===
    function addLog(type, message) {
        setLogs((prev) => [
            ...prev,
            { timestamp: getCurrentTimestamp(), type, message },
        ]);
    }

    function getCurrentTimestamp() {
        return new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    }

    function getUserInitial(avatar) {
        return avatar?.username?.charAt(0)?.toUpperCase() || "?";
    }

    return (
        <div className="game-container bg-gradient-to-b from-blue-400 to-blue-600 p-2 rounded-lg shadow-xl max-w-4xl mx-auto font-sans relative overflow-hidden border-4 border-yellow-500">
            {/* Decorative elements */}
            <div className="absolute -top-16 -left-16 w-32 h-32 bg-yellow-300 rounded-full opacity-20"></div>
            <div className="absolute -bottom-16 -right-16 w-32 h-32 bg-yellow-300 rounded-full opacity-20"></div>

            {/* Game title */}
            <div className="text-center mb-2">
                <h1 className="text-3xl font-bold text-yellow-300 drop-shadow-md transform rotate-2 mb-1">Tower Clash</h1>
                <div className="w-32 h-1 bg-yellow-400 mx-auto rounded-full mb-2"></div>
            </div>

            {/* STATS BAR */}
            <div className="stats-bar flex justify-between items-center p-2 bg-gradient-to-r from-blue-900 to-blue-800 rounded-lg shadow-md mb-2">
                {/* OPPONENT STATS */}
                <div className="opponent-stats flex items-center">
                    <div className="opponent-avatar relative">
                        <div className="bg-red-500 rounded-full w-12 h-12 flex items-center justify-center text-white font-bold border-2 border-red-700 shadow-md transform hover:scale-105 transition-transform">
                            {getUserInitial(opponent.user)}
                        </div>
                        <div className="absolute -bottom-1 -right-1 bg-red-700 text-white text-xs rounded-full w-6 h-6 flex items-center justify-center border border-yellow-400">
                            {opponent.user?.level || 0}
                        </div>
                    </div>
                    <div className="stat-column ml-2">
                        <div className="stat">
                            <div className="stat-value name text-yellow-300 font-bold text-lg drop-shadow-md">
                                {opponent.user?.username || "Waiting..."}
                            </div>
                        </div>
                    </div>
                </div>

                {/* TURN DISPLAY */}
                <div className="turn-display text-center transform hover:scale-105 transition-transform">
                    <div className={`font-bold text-lg px-4 py-1 rounded-full ${game.playerTurn === user.user?.username
                        ? "bg-green-600 text-white animate-pulse"
                        : "bg-red-600 text-white"
                        }`}>
                        {game.playerTurn === user.user?.username ? "YOUR TURN" : "OPPONENT'S TURN"}
                    </div>
                    <button
                        className={`skip-btn mt-1 px-4 py-1 rounded-full font-semibold transition-all transform hover:scale-105 ${game.playerTurn === user.user?.username
                            ? "bg-yellow-400 text-blue-900 border-2 border-yellow-500"
                            : "bg-gray-500 text-white opacity-50 cursor-not-allowed"
                            }`}
                        disabled={game.playerTurn !== user.user?.username}
                        onClick={handleSkipTurn}
                    >
                        Skip + 2 Mana
                    </button>
                </div>

                {/* PLAYER STATS */}
                <div className="player-stats flex items-center">
                    <div className="user-avatar relative">
                        <div className="bg-blue-500 rounded-full w-12 h-12 flex items-center justify-center text-white font-bold border-2 border-blue-700 shadow-md transform hover:scale-105 transition-transform">
                            {getUserInitial(user.user)}
                        </div>
                        <div className="absolute -bottom-1 -right-1 bg-blue-700 text-white text-xs rounded-full w-6 h-6 flex items-center justify-center border border-yellow-400">
                            {user.user?.level || 0}
                        </div>
                    </div>
                    <div className="stat-column ml-2">
                        <div className="stat">
                            <div className="stat-value name text-yellow-300 font-bold text-lg drop-shadow-md">
                                {user.user?.username || "Loading..."}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* BATTLEFIELD */}
            <div className="battle-container bg-gradient-to-b from-green-500 to-green-600 rounded-lg shadow-inner min-h-[400px] border-2 border-green-700 overflow-hidden relative">
                {/* Decorative elements */}
                <div className="absolute inset-0 bg-[url('https://cdnjs.cloudflare.com/ajax/libs/simple-icons/3.0.1/simpleicons.svg')] opacity-5"></div>
                <div className="absolute top-1/2 left-0 w-full h-2 bg-gray-800 opacity-20"></div>

                <div className="battlefield flex flex-col h-full">
                    {/* Opponent Side */}
                    <div className="player-side opponent p-4 relative">
                        {/* King Tower */}
                        <div className={`tower king mx-auto mb-6 relative ${game.opponentHealth.king <= 0 ? 'grayscale' : ''
                            }`} onClick={() => selectTarget("king")}>
                            <div className="absolute -top-6 left-1/2 transform -translate-x-1/2 text-red-500 font-bold text-lg opacity-0 animate-bounce transition-opacity" id="damage-indicator-king">
                                -{game.selectedTroop?.atk || 0}
                            </div>
                            <div className="tower-content bg-gradient-to-b from-red-400 to-red-600 p-3 rounded-lg border-4 border-red-700 shadow-lg transform hover:scale-105 transition-transform cursor-pointer">
                                <div className="tower-icon text-center text-4xl drop-shadow-md">👑</div>
                                <div className="tower-hp mt-2">
                                    <div className="hp-bar bg-gray-700 w-full h-4 rounded-full shadow-inner overflow-hidden border border-gray-800">
                                        <div
                                            className="hp-fill bg-gradient-to-r from-green-500 to-green-400 h-full rounded-full transition-all duration-500"
                                            style={{
                                                width: `${Math.max(0, (game.opponentHealth.king / game.opponentShield.king) * 100)}%`
                                            }}
                                        />
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* Guard Towers */}
                        <div className="tower-container flex justify-center space-x-12">
                            <div
                                className={`tower guard ${game.opponentHealth.guard1 <= 0 ? 'grayscale' : ''}`}
                                onClick={() => selectTarget("guard1")}
                            >
                                <div className="absolute -top-6 left-1/2 transform -translate-x-1/2 text-red-500 font-bold text-lg opacity-0 animate-bounce transition-opacity" id="damage-indicator-guard1">
                                    -{game.selectedTroop?.atk || 0}
                                </div>
                                <div className="tower-content bg-gradient-to-b from-red-300 to-red-500 p-2 rounded-lg border-3 border-red-600 shadow-lg transform hover:scale-105 transition-transform cursor-pointer">
                                    <div className="tower-icon text-center text-3xl drop-shadow-md">🛡️</div>
                                    <div className="tower-hp mt-2">
                                        <div className="hp-bar bg-gray-700 w-full h-3 rounded-full shadow-inner overflow-hidden border border-gray-800">
                                            <div
                                                className="hp-fill bg-gradient-to-r from-green-500 to-green-400 h-full rounded-full transition-all duration-500"
                                                style={{
                                                    width: `${Math.max(0, (game.opponentHealth.guard1 / game.opponentShield.guard1) * 100)}%`
                                                }}
                                            />
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div
                                className={`tower guard ${game.opponentHealth.guard2 <= 0 ? 'grayscale' : ''}`}
                                onClick={() => selectTarget("guard2")}
                            >
                                <div className="absolute -top-6 left-1/2 transform -translate-x-1/2 text-red-500 font-bold text-lg opacity-0 animate-bounce transition-opacity" id="damage-indicator-guard2">
                                    -{game.selectedTroop?.atk || 0}
                                </div>
                                <div className="tower-content bg-gradient-to-b from-red-300 to-red-500 p-2 rounded-lg border-3 border-red-600 shadow-lg transform hover:scale-105 transition-transform cursor-pointer">
                                    <div className="tower-icon text-center text-3xl drop-shadow-md">🛡️</div>
                                    <div className="tower-hp mt-2">
                                        <div className="hp-bar bg-gray-700 w-full h-3 rounded-full shadow-inner overflow-hidden border border-gray-800">
                                            <div
                                                className="hp-fill bg-gradient-to-r from-green-500 to-green-400 h-full rounded-full transition-all duration-500"
                                                style={{
                                                    width: `${Math.max(0, (game.opponentHealth.guard2 / game.opponentShield.guard2) * 100)}%`
                                                }}
                                            />
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Player Side */}
                    <div className="player-side p-4 mt-4 relative">
                        {/* Guard Towers */}
                        <div className="tower-container flex justify-center space-x-12">
                            <div className={`tower guard ${game.playerHealth.guard1 <= 0 ? 'grayscale' : ''}`}>
                                <div className="tower-content bg-gradient-to-b from-blue-300 to-blue-500 p-2 rounded-lg border-3 border-blue-600 shadow-lg">
                                    <div className="tower-icon text-center text-3xl drop-shadow-md">🛡️</div>
                                    <div className="tower-hp mt-2">
                                        <div className="hp-bar bg-gray-700 w-full h-3 rounded-full shadow-inner overflow-hidden border border-gray-800">
                                            <div
                                                className="hp-fill bg-gradient-to-r from-green-500 to-green-400 h-full rounded-full transition-all duration-500"
                                                style={{
                                                    width: `${Math.max(0, (game.playerHealth.guard1 / game.playerShield.guard1) * 100)}%`
                                                }}
                                            />
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div className={`tower guard ${game.playerHealth.guard2 <= 0 ? 'grayscale' : ''}`}>
                                <div className="tower-content bg-gradient-to-b from-blue-300 to-blue-500 p-2 rounded-lg border-3 border-blue-600 shadow-lg">
                                    <div className="tower-icon text-center text-3xl drop-shadow-md">🛡️</div>
                                    <div className="tower-hp mt-2">
                                        <div className="hp-bar bg-gray-700 w-full h-3 rounded-full shadow-inner overflow-hidden border border-gray-800">
                                            <div
                                                className="hp-fill bg-gradient-to-r from-green-500 to-green-400 h-full rounded-full transition-all duration-500"
                                                style={{
                                                    width: `${Math.max(0, (game.playerHealth.guard2 / game.playerShield.guard2) * 100)}%`
                                                }}
                                            />
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* King Tower */}
                        <div className={`tower king mx-auto mt-6 ${game.playerHealth.king <= 0 ? 'grayscale' : ''}`}>
                            <div className="tower-content bg-gradient-to-b from-blue-400 to-blue-600 p-3 rounded-lg border-4 border-blue-700 shadow-lg">
                                <div className="tower-icon text-center text-4xl drop-shadow-md">👑</div>
                                <div className="tower-hp mt-2">
                                    <div className="hp-bar bg-gray-700 w-full h-4 rounded-full shadow-inner overflow-hidden border border-gray-800">
                                        <div
                                            className="hp-fill bg-gradient-to-r from-green-500 to-green-400 h-full rounded-full transition-all duration-500"
                                            style={{
                                                width: `${Math.max(0, (game.playerHealth.king / game.playerShield.king) * 100)}%`
                                            }}
                                        />
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* MANA BAR */}
            <div className="mana-container bg-gradient-to-r from-blue-900 to-blue-800 p-2 rounded-lg border-2 border-blue-700 my-2 shadow-md">
                <div className="flex items-center justify-between mb-1">
                    <div className="text-lg text-yellow-400 font-bold flex items-center">
                        <span className="text-xl mr-1">⚡</span> MANA
                    </div>
                    <div className="text-white font-bold">{game.playerMana}/{game.maxMana}</div>
                </div>
                <div className="mana-bar bg-gray-800 h-6 rounded-full shadow-inner overflow-hidden border border-gray-900 flex">
                    {Array.from({ length: game.maxMana }).map((_, i) => (
                        <div
                            key={i}
                            className={`mana-segment flex-1 h-full border-r border-gray-700 last:border-r-0 transition-all ${i < game.playerMana
                                ? 'bg-gradient-to-r from-blue-500 to-blue-400'
                                : ''
                                }`}
                        />
                    ))}
                </div>
            </div>

            {/* TROOP SELECTION */}
            <div className="troops-container bg-gradient-to-r from-blue-900 to-blue-800 p-4 rounded-lg mt-2 shadow-md border-2 border-blue-700">
                <div className="section-header">
                    <h3 className="text-xl font-bold text-center mb-3 text-yellow-400 drop-shadow-md">TROOPS</h3>
                </div>

                <div className="troop-selection flex justify-center space-x-4">
                    {Object.entries(game.troops).map(([troopName, troop], index) => (
                        <div
                            key={index}
                            className={`troop ${game.selectedTroop?.name === troopName
                                ? 'border-4 border-yellow-400 bg-yellow-100 transform scale-105'
                                : 'border-2 border-gray-400 bg-white'
                                } ${game.playerMana < troop.mana
                                    ? 'opacity-50 grayscale'
                                    : 'hover:scale-105'
                                } rounded-lg shadow-lg p-2 cursor-pointer transition-all duration-200`}
                            onClick={() => selectTroop(troopName)}
                        >
                            <div className="troop-banner bg-gradient-to-r from-blue-600 to-blue-500 rounded-t-md px-2 py-1 -mt-2 -mx-2 mb-1 text-center">
                                <div className="troop-name text-white font-bold drop-shadow-md">{troopName}</div>
                            </div>
                            <div className="troop-icon text-center text-4xl mb-1">{troop.icon}</div>
                            <div className="flex justify-between items-center mb-1">
                                <div className="troop-mana-cost bg-blue-500 text-white font-bold flex items-center rounded-full px-2 border-2 border-blue-600">
                                    <span className="text-yellow-300 mr-1">⚡</span> {troop.mana}
                                </div>
                                {troop.atk && (
                                    <div className="troop-damage bg-red-500 text-white font-bold flex items-center rounded-full px-2 border-2 border-red-600">
                                        <span className="text-yellow-300 mr-1">💥</span> {troop.atk}
                                    </div>
                                )}
                            </div>
                            <div className="ability-description text-xs text-center bg-gray-100 p-1 rounded">
                                {troop.description}
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            {/* GAME LOG */}
            <div className="log-container p-2 bg-gradient-to-r from-gray-900 to-gray-800 max-h-28 overflow-y-auto mt-2 rounded-lg border-2 border-gray-700 shadow-inner">
                <h4 className="text-center text-yellow-400 font-bold mb-1 text-sm">BATTLE LOG</h4>
                {logs.map((log, index) => (
                    <div
                        className={`log-entry text-sm mb-1 px-2 py-1 rounded ${index === 0 ? 'bg-gray-800 animate-pulse' : ''}`}
                        key={index}
                    >
                        <span className="timestamp font-mono text-gray-400">[{log.timestamp}]</span>
                        <span className={`type font-bold ${log.type === "SYSTEM" ? "text-blue-400" :
                            log.type === "ACTION" ? "text-green-400" :
                                log.type === "TOWER" ? "text-red-400" : ""
                            }`}> {log.type}:</span>
                        <span className="message text-white"> {log.message}</span>
                    </div>
                ))}
            </div>

            {/* NOTIFICATION */}
            {notification.show && (
                <div className="notification fixed top-5 left-0 right-0 mx-auto max-w-md bg-yellow-400 px-4 py-2 text-center rounded-full shadow-lg border-2 border-yellow-500 animate-bounce">
                    <span className="font-bold text-blue-900">{notification.message}</span>
                </div>
            )}

            {/* GAME OVER MODAL */}
            {game.gameOver && (
                <div className="game-over-modal fixed inset-0 bg-black bg-opacity-70 flex items-center justify-center z-50">
                    <div className="modal-content bg-gradient-to-b from-blue-800 to-blue-900 rounded-lg shadow-xl p-6 max-w-md w-full border-4 border-yellow-500 transform scale-105 animate-pulse">
                        {/* Crown decoration */}
                        <div className="crown-decoration absolute -top-10 left-1/2 transform -translate-x-1/2 text-6xl">
                            {game.winner === user.user?.username ? "👑" : "☠️"}
                        </div>

                        <h2 className={`modal-title text-3xl font-bold mb-4 text-center ${game.winner === user.user?.username ? "text-yellow-400" : "text-red-400"
                            }`}>
                            {game.winner === user.user?.username ? "VICTORY!" : "DEFEAT"}
                        </h2>

                        <div className="modal-body text-center">
                            <p className="text-white text-lg mb-3">
                                {game.winner === user.user?.username
                                    ? "You have conquered your opponent's kingdom!"
                                    : "Your kingdom has fallen to the enemy!"}
                            </p>
                            <div className="exp-gain text-yellow-300 font-bold text-2xl mt-3 animate-pulse">
                                {game.winner === user.user?.username ? "+100 XP" : "+25 XP"}
                            </div>

                            <div className="stats-summary mt-6 bg-blue-950 p-4 rounded-lg border-2 border-blue-700">
                                <h3 className="text-yellow-400 font-bold text-lg mb-2 text-center">BATTLE STATS</h3>
                                <div className="stats-grid grid grid-cols-2 gap-3">
                                    <div className="stat-item bg-blue-800 p-2 rounded-lg border border-blue-600">
                                        <div className="stat-icon text-center text-2xl">⏱️</div>
                                        <div className="stat-label text-center text-white text-sm">Turns Played</div>
                                        <div className="stat-value text-center text-yellow-300 font-bold text-xl">
                                            {game.stats?.turns || 0}
                                        </div>
                                    </div>

                                    <div className="stat-item bg-blue-800 p-2 rounded-lg border border-blue-600">
                                        <div className="stat-icon text-center text-2xl">💥</div>
                                        <div className="stat-label text-center text-white text-sm">Damage Dealt</div>
                                        <div className="stat-value text-center text-yellow-300 font-bold text-xl">
                                            {game.stats?.damageDealt || 0}
                                        </div>
                                    </div>

                                    <div className="stat-item bg-blue-800 p-2 rounded-lg border border-blue-600">
                                        <div className="stat-icon text-center text-2xl">⚡</div>
                                        <div className="stat-label text-center text-white text-sm">Mana Spent</div>
                                        <div className="stat-value text-center text-yellow-300 font-bold text-xl">
                                            {game.stats?.manaSpent || 0}
                                        </div>
                                    </div>

                                    <div className="stat-item bg-blue-800 p-2 rounded-lg border border-blue-600">
                                        <div className="stat-icon text-center text-2xl">✨</div>
                                        <div className="stat-label text-center text-white text-sm">Critical Hits</div>
                                        <div className="stat-value text-center text-yellow-300 font-bold text-xl">
                                            {game.stats?.criticalHits || 0}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div className="modal-buttons mt-6 text-center">
                            <button
                                className="play-again-btn bg-gradient-to-r from-yellow-400 to-yellow-500 text-blue-900 px-8 py-3 rounded-full font-bold text-lg border-4 border-yellow-600 shadow-lg transform hover:scale-105 transition-transform"
                                onClick={handlePlayAgain}
                            >
                                PLAY AGAIN
                            </button>
                        </div>
                    </div>
                </div>
            )}

        </div>
    );
}