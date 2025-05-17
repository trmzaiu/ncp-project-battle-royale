import { useEffect, useState, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { useWebSocketContext } from "../context/WebSocketContext";

export default function Game() {
    const navigate = useNavigate();
    const { sendMessage, subscribe } = useWebSocketContext();
    const damageTimeoutRef = useRef(null);
    const healTimeoutRef = useRef(null);

    const [user, setUser] = useState({});
    const [opponent, setOpponent] = useState({});
    const [isGameInitialized, setIsGameInitialized] = useState(false);

    const [damagePopup, setDamagePopup] = useState({
        targetId: null,
        amount: 0,
        isOpponent: false,
        visible: false,
        isCrit: false,
    });

    const [healPopup, setHealPopup] = useState({
        target: null,
        amount: 0,
        isOpponent: false,
        visible: false,
    });

    const [game, setGame] = useState(getInitialGameState());
    const [notification, setNotification] = useState({
        show: false,
        message: "",
    });

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
            winner: null,
        };
    }

    // === Notification Helper ===
    const showNotification = (message) => {
        setNotification({ show: true, message });
        setTimeout(
            () => setNotification((prev) => ({ ...prev, show: false })),
            4000
        );
    };

    // === Effect: Initial Setup & WebSocket Subscription ===
    useEffect(() => {
        if (!localStorage.getItem("session_id")) {
            showNotification("Session expired. Redirecting to login...");
            setTimeout(() => navigate("/auth"), 1500);
            return;
        }

        if (!localStorage.getItem("room_id")) {
            showNotification("Room not found. Redirecting to lobby...");
            setTimeout(() => navigate("/lobby"), 1500);
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

        return () => {
            unsubscribe();
            if (damageTimeoutRef.current) clearTimeout(damageTimeoutRef.current);
            if (healTimeoutRef.current) clearTimeout(healTimeoutRef.current);
        };

    }, [subscribe, sendMessage, navigate]);

    // === Message Handler ===
    const handleMessage = (res) => {
        switch (res.type) {
            case "game_response":
                console.log("Game Response:", res);
                if (res.success) {
                    handleSetGameState(res.data);
                } else {
                    showNotification(res.error || "Failed to get game data");
                }
                break;

            case "attack_response":
                console.log("Attack Response:", res);
                if (res.success) {
                    handleAttack(res);
                } else if (!res.success && res.data.attacker.user.username === localStorage.getItem("username")) {
                    showNotification(res.message);
                }
                break;

            case "heal_response":
                console.log("Heal Response:", res);
                if (res.success) {
                    handleHeal(res);
                }
                break;

            case "skip_turn_response":
                res.success && handleSkipTurn(res);
                break;

            case "game_over_response":
                res.success && handleGameOver(res);
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
            playerShield: extractMaxHP(userData.towers),
            opponentShield: extractMaxHP(opponentData.towers),
            troops: troops.reduce((acc, troop) => {
                acc[troop.name] = troop;
                return acc;
            }, {}),
            selectedTroop: null,
            selectedTarget: null,
            playerTurn: turn,
            gameOver: false,
            winner: null
        });

        if (turn === localStorage.getItem("username")) {
            showNotification("Your turn.");
        } else {
            showNotification("Opponent's turn.");
        }

        setIsGameInitialized(true);
    };

    const extractMaxHP = (towers) => ({
        king: towers.king.max_hp,
        guard1: towers.guard1.max_hp,
        guard2: towers.guard2.max_hp,
    });

    const extractHP = (towers) => ({
        king: towers.king.hp,
        guard1: towers.guard1.hp,
        guard2: towers.guard2.hp,
    });

    // === Skip Turn ===
    const skipTurn = () => {
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
    };

    // === Select Troop ===
    const selectTroop = (troopName) => {
        if (game.playerTurn !== user.user?.username)
            return showNotification("Not your turn.");

        const troop = game.troops[troopName];
        if (game.playerMana < troop.mana)
            return showNotification("Not enough mana.");

        if (troop.name === "Queen") {
            sendMessage({
                type: "heal",
                data: {
                    troop: troop.name,
                    room_id: localStorage.getItem("room_id"),
                    username: user.user?.username,
                },
            });

            return;
        }

        setGame((prev) => ({ ...prev, selectedTroop: troop }));
    };

    // === Select Target & Attack ===
    const selectTarget = (target) => {
        const { selectedTroop, playerMana, playerTurn } = game;
        const currentUser = user.user?.username;

        if (playerTurn !== currentUser || !selectedTroop)
            return showNotification("Invalid action.");
        if (playerMana < selectedTroop.mana)
            return showNotification("Not enough mana.");

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

    // === Handle Set Game Response ===
    const handleSetGameState = (msg) => {
        const { turn, user, opponent } = msg;
        setUser(user);
        setOpponent(opponent);

        if (!isGameInitialized) {
            initializeGame(
                turn,
                user.troops,
                user,
                opponent
            );
        }
    }

    // === Handle Attack Response ===
    const handleAttack = (msg) => {
        const { attacker, defender, damage, target, isDestroyed, turn, troop, isCrit } =
            msg.data;
        const isMe = attacker.user.username === localStorage.getItem("username");
        const targetId = (isMe ? "opponent-" : "player-") + target;

        if (damageTimeoutRef.current) {
            clearTimeout(damageTimeoutRef.current);
            // Immediately hide any existing popup before showing the new one
            setDamagePopup(prev => ({ ...prev, visible: false }));
        }

        setGame((prev) => {
            const newState = { ...prev };

            if (isMe) {
                setUser(attacker);
                setOpponent(defender);
                newState.playerMana = attacker.mana;
                newState.opponentHealth[target] = defender.towers[target].hp;
                showNotification(
                    `Your ${troop} dealt ${damage} damage to opponent's ${target}.`
                );
            } else {
                setUser(defender);
                setOpponent(attacker);
                newState.playerMana = defender.mana;
                newState.playerHealth[target] = defender.towers[target].hp;
                showNotification(
                    `Opponent's ${troop} dealt ${damage} damage to your ${target}.`
                );
            }

            newState.playerTurn = turn;
            return newState;
        });

        setTimeout(() => {
            setDamagePopup({
                targetId,
                amount: damage,
                isOpponent: isMe,
                visible: true,
                crit: isCrit
            });

            // Set new timeout and store its ID
            damageTimeoutRef.current = setTimeout(() => {
                setDamagePopup(prev => ({ ...prev, visible: false }));
                damageTimeoutRef.current = null;
            }, 2000);
        }, 50);

        if (isDestroyed) showNotification(`Tower ${target} has been destroyed!`);
        showNotification(
            turn === user?.user?.username
                ? "Your turn."
                : "Waiting for opponent's turn..."
        );
    };

    // === Handle Heal Response ===
    const handleHeal = (msg) => {
        const { turn, player, opponent, troop, healedTower, healAmount } = msg.data;
        const isMe = player.user.username === localStorage.getItem("username");
        const targetId = (isMe ? "player-" : "opponent-") + healedTower.type;

        if (healTimeoutRef.current) {
            clearTimeout(healTimeoutRef.current);
            // Immediately hide any existing popup before showing the new one
            setHealPopup(prev => ({ ...prev, visible: false }));
        }

        setGame((prev) => {
            const newState = { ...prev };

            if (isMe) {
                setUser(player);
                setOpponent(opponent);
                newState.playerMana = player.mana;
                newState.playerHealth[healedTower.type] = healedTower.hp;
                showNotification(
                    `Your ${troop} healed your ${healedTower.type} for ${healAmount} HP.`
                );
            } else {
                setUser(opponent);
                setOpponent(player);
                newState.playerMana = opponent.mana;
                newState.opponentHealth[healedTower.type] = healedTower.hp;
                showNotification(
                    `Opponent's ${troop} healed their ${healedTower.type} for ${healAmount} HP.`
                );
            }

            newState.playerTurn = turn;
            return newState;
        });

        setTimeout(() => {
            setHealPopup({
                targetId,
                amount: healAmount,
                isOpponent: !isMe,
                visible: true,
            });

            // Set new timeout and store its ID
            healTimeoutRef.current = setTimeout(() => {
                setHealPopup(prev => ({ ...prev, visible: false }));
                healTimeoutRef.current = null;
            }, 2000);
        }, 50);

        showNotification(
            turn === localStorage.getItem("username")
                ? "Your turn."
                : "Waiting for opponent's turn..."
        );
    };


    // === Handle Skip Turn Response
    const handleSkipTurn = (msg) => {
        const { turn, player1, player2 } = msg.data;

        const currentUsername = localStorage.getItem("username");
        const updatedSelf = player1?.user?.username === currentUsername ? player1 : player2;

        setGame((prev) => ({
            ...prev,
            playerTurn: turn,
            playerMana: updatedSelf?.mana ?? prev.playerMana,
        }));

        showNotification("Turn skipped.");
    };


    // === Handle Game Over ===
    const handleGameOver = (res) => {
        setGame((prev) => ({ ...prev, gameOver: true, winner: res.data.winner }));
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
        localStorage.removeItem("room_id");
        setTimeout(() => navigate("/lobby"), 1500);
    };

    // Tower component for reusability 
    const Tower = ({ id, type, health, maxHealth, isOpponent, onClick, disabled }) => {
        const towerIcon = type === "king" ? "👑" : "🛡️";
        const towerSize = type === "king" ? "text-4xl" : "text-3xl";
        const borderWidth = type === "king" ? "border-4" : "border-3";
        const borderColor = isOpponent ? "border-red-700" : "border-blue-700";

        let bgGradient;
        if (isOpponent) {
            bgGradient = type === "king"
                ? "bg-gradient-to-b from-red-400 to-red-600"
                : "bg-gradient-to-b from-red-300 to-red-500";
        } else {
            bgGradient = type === "king"
                ? "bg-gradient-to-b from-blue-400 to-blue-600"
                : "bg-gradient-to-b from-blue-300 to-blue-500";
        }

        return (
            <div
                className={`relative tower ${type} ${health <= 0 ? "grayscale" : ""} w-full h-full flex items-center justify-center ${!disabled ? "cursor-pointer" : ""}`}
                onClick={disabled ? undefined : onClick}
            >
                {damagePopup?.visible && damagePopup.targetId === id && (
                    <div
                        className={`absolute ${isOpponent ? "-bottom-10" : "-top-10"} pointer-events-none`}
                    >
                        <div className={`
                        flex justify-center items-center
                        ${damagePopup.crit
                                ? "text-yellow-300 animate-clash-crit-popup"
                                : "text-red-500 animate-clash-damage-popup"}
                    `}>
                            {/* Clash Royale style outline effect */}
                            <div className="relative">
                                <span className={`absolute font-extrabold ${damagePopup.crit ? "text-4xl" : "text-3xl"} text-white opacity-25 -left-0.5`}>-{damagePopup.amount}</span>
                                <span className={`absolute font-extrabold ${damagePopup.crit ? "text-4xl" : "text-3xl"} text-white opacity-25 -right-0.5`}>-{damagePopup.amount}</span>
                                <span className={`absolute font-extrabold ${damagePopup.crit ? "text-4xl" : "text-3xl"} text-white opacity-25 -top-0.5`}>-{damagePopup.amount}</span>
                                <span className={`absolute font-extrabold ${damagePopup.crit ? "text-4xl" : "text-3xl"} text-white opacity-25 -bottom-0.5`}>-{damagePopup.amount}</span>
                                {/* Actual text */}
                                <span className={`relative font-extrabold ${damagePopup.crit ? "text-4xl" : "text-3xl"}`}>
                                    -{damagePopup.amount}
                                </span>
                            </div>
                            {damagePopup.crit && (
                                <div className="absolute w-16 h-16 bg-yellow-400 rounded-full opacity-20 animate-clash-crit-burst" />
                            )}
                        </div>
                    </div>
                )}

                {/* Clash Royale Style Heal Popup */}
                {healPopup?.visible && healPopup.targetId === id && (
                    <div
                        className={`absolute ${isOpponent ? "-bottom-10" : "-top-10"} pointer-events-none`}
                    >
                        <div className="flex justify-center items-center text-green-700 animate-clash-heal-popup">
                            {/* Clash Royale style outline effect */}
                            <div className="relative">
                                <span className="absolute font-extrabold text-3xl text-white opacity-25 -left-0.5">+{healPopup.amount}</span>
                                <span className="absolute font-extrabold text-3xl text-white opacity-25 -right-0.5">+{healPopup.amount}</span>
                                <span className="absolute font-extrabold text-3xl text-white opacity-25 -top-0.5">+{healPopup.amount}</span>
                                <span className="absolute font-extrabold text-3xl text-white opacity-25 -bottom-0.5">+{healPopup.amount}</span>
                                {/* Actual text */}
                                <span className="relative font-extrabold text-3xl">
                                    +{healPopup.amount}
                                </span>
                            </div>
                            <div className="absolute w-12 h-12 bg-green-400 rounded-full opacity-20 animate-clash-heal-burst" />
                        </div>
                    </div>
                )}

                <div className={`tower-content ${bgGradient} p-2 rounded-lg ${borderWidth} ${borderColor} shadow-lg ${!disabled ? "transform hover:scale-105 transition-transform" : ""} w-full h-full flex flex-col items-center justify-center`}>
                    <div className={`tower-icon text-center ${towerSize} drop-shadow-md`}>
                        {towerIcon}
                    </div>
                    <div className="tower-hp mt-2 w-full">
                        <div className="hp-bar bg-gray-700 w-full h-3 rounded-full shadow-inner overflow-hidden border border-gray-800">
                            <div
                                className={`hp-fill bg-gradient-to-r ${health <= maxHealth / 3 ? "from-red-500 to-red-400" : "from-green-500 to-green-400"} h-full rounded-full transition-all duration-500`}
                                style={{
                                    width: `${Math.max(0, (health / maxHealth) * 100)}%`,
                                }}
                            />
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    return (
        <div className="min-h-screen bg-gradient-to-b from-blue-900 via-blue-500 to-blue-900">
            <div className="game-container bg-gradient-to-b from-blue-400 to-blue-600 p-2 rounded-lg shadow-xl max-w-4xl mx-auto font-sans relative overflow-hidden border-4 border-yellow-500">
                {/* Decorative elements */}
                <div className="absolute -top-16 -left-16 w-32 h-32 bg-yellow-300 rounded-full opacity-20"></div>
                <div className="absolute -bottom-16 -right-16 w-32 h-32 bg-yellow-300 rounded-full opacity-20"></div>

                {/* Game title */}
                <div className="text-center mb-2">
                    <h1 className="text-3xl font-bold text-yellow-300 drop-shadow-md transform rotate-2 mb-1">
                        ROYAKA
                    </h1>
                    <div className="w-32 h-1 bg-yellow-400 mx-auto rounded-full mb-2"></div>
                </div>

                {/* STATS BAR */}
                <div className="stats-bar flex justify-between items-center p-2 bg-gradient-to-r from-blue-900 to-blue-800 rounded-lg shadow-md mb-2">
                    {/* OPPONENT STATS */}
                    <div className="opponent-stats flex items-center">
                        <div className="opponent-avatar relative">
                            <img
                                src={opponent.user?.avatar}
                                alt="avatar"
                                className="w-12 h-12 rounded-full border-2 border-red-700 shadow-md transform hover:scale-105 transition-transform object-cover"
                            />
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
                        <div
                            className={`font-bold text-lg px-4 py-1 rounded-full ${game.playerTurn === user.user?.username
                                ? "bg-green-600 text-white animate-pulse"
                                : "bg-red-600 text-white"
                                }`}
                        >
                            {game.playerTurn === user.user?.username
                                ? "YOUR TURN"
                                : "OPPONENT'S TURN"}
                        </div>
                    </div>
                </div>

                {/* BATTLEFIELD - GRID LAYOUT */}
                <div className="battle-container bg-gradient-to-b from-green-500 to-green-600 rounded-lg shadow-inner border-2 border-green-700 overflow-hidden relative">
                    {/* Grid background */}
                    <div className="absolute inset-0 grid grid-cols-10 grid-rows-6">
                        {Array.from({ length: 60 }).map((_, i) => {
                            const row = Math.floor(i / 10);
                            const col = i % 10;
                            // If row + col is even, color A; else color B
                            const isEven = (row + col) % 2 === 0;
                            return (
                                <div
                                    key={i}
                                    className={`border border-green-500 aspect-square ${isEven ? "bg-green-400 bg-opacity-20" : "bg-green-500 bg-opacity-30"
                                        }`}
                                ></div>
                            );
                        })}
                    </div>


                    {/* Battlefield grid layout - 10 columns x 6 rows */}
                    <div className="grid-battlefield grid grid-cols-10 grid-rows-6 relative">
                        {/* Row 1 - Opponent King (spans 2 columns) */}
                        <div className="col-start-5 col-span-2 relative ">
                            {" "}
                            {/* Starts at column 5, spans 2 columns */}
                            <Tower
                                type="king"
                                id="opponent-king"
                                health={game.opponentHealth.king}
                                maxHealth={game.opponentShield.king}
                                isOpponent={true}
                                onClick={() => selectTarget("king")}
                                disabled={!game.selectedTroop}
                                className="w-full h-full"
                            />
                        </div>

                        {[...Array(4)].map((_, i) => (
                            <div key={`r1-fill-${i}`} className="cell"></div>
                        ))}

                        {/* Row 2 - Opponent guards */}
                        {Array.from({ length: 10 }).map((_, i) => (
                            <div key={`r2-${i}`} className="cell">
                                {i === 2 && (
                                    <Tower
                                        type="guard"
                                        id="opponent-guard1"
                                        health={game.opponentHealth.guard1}
                                        maxHealth={game.opponentShield.guard1}
                                        isOpponent={true}
                                        onClick={() => selectTarget("guard1")}
                                        disabled={!game.selectedTroop}
                                    />
                                )}
                                {i === 7 && (
                                    <Tower
                                        type="guard"
                                        id="opponent-guard2"
                                        health={game.opponentHealth.guard2}
                                        maxHealth={game.opponentShield.guard2}
                                        isOpponent={true}
                                        onClick={() => selectTarget("guard2")}
                                        disabled={!game.selectedTroop}
                                    />
                                )}
                            </div>
                        ))}

                        {/* Rows 3-4 - Middle empty rows */}
                        {Array.from({ length: 2 }).map((_, rowIndex) =>
                            Array.from({ length: 10 }).map((_, colIndex) => (
                                <div
                                    key={`r${rowIndex + 3}-${colIndex}`}
                                    className="cell"
                                ></div>
                            ))
                        )}

                        {/* Row 5 - Player guards */}
                        {Array.from({ length: 10 }).map((_, i) => (
                            <div key={`r5-${i}`} className="cell">
                                {i === 2 && (
                                    <Tower
                                        type="guard"
                                        id="player-guard1"
                                        health={game.playerHealth.guard1}
                                        maxHealth={game.playerShield.guard1}
                                        isOpponent={false}
                                        disabled={true}
                                    />
                                )}
                                {i === 7 && (
                                    <Tower
                                        type="guard"
                                        id="player-guard2"
                                        health={game.playerHealth.guard2}
                                        maxHealth={game.playerShield.guard2}
                                        isOpponent={false}
                                        disabled={true}
                                    />
                                )}
                            </div>
                        ))}

                        {/* Row 6 - Player King (spans 2 columns) */}
                        <div className="col-start-5 col-span-2 relative">
                            {" "}
                            {/* Starts at column 5, spans 2 columns */}
                            <Tower
                                type="king"
                                id="player-king"
                                health={game.playerHealth.king}
                                maxHealth={game.playerShield.king}
                                isOpponent={false}
                                disabled={true}
                                className="w-[calc(200%+8px)] h-full -ml-1" /* 200% + gap */
                            />
                        </div>

                        {/* Fill remaining cells in row 6 */}
                        {[...Array(4), ...Array(4)].map((_, i) => (
                            <div key={`r6-fill-${i}`} className="cell"></div>
                        ))}
                    </div>

                    {/* Target indicators */}
                    {game.selectedTroop && (
                        <div className="target-indicators absolute top-0 left-0 w-full h-full pointer-events-none">
                            <div className="text-center text-white font-bold text-lg absolute top-1/3 left-1/2 transform -translate-x-1/2 -translate-y-1/2 bg-blue-800 bg-opacity-70 px-4 py-2 rounded-full">
                                🎯 Pick a Target!
                            </div>
                        </div>
                    )}
                </div>

                {/* MANA BAR */}
                <div className="mana-container bg-gradient-to-r from-blue-900 to-blue-800 p-2 rounded-lg border-2 border-blue-700 my-2 shadow-md">
                    <div className="flex items-center justify-between mb-1">
                        <div className="text-lg text-yellow-400 font-bold flex items-center">
                            <span className="text-xl mr-1">⚡</span> MANA
                        </div>
                        <div className="text-white font-bold">
                            {game.playerMana}/{game.maxMana}
                        </div>
                    </div>
                    <div className="mana-bar bg-gray-800 h-6 rounded-full shadow-inner overflow-hidden border border-gray-900 flex">
                        {Array.from({ length: game.maxMana }).map((_, i) => (
                            <div
                                key={i}
                                className={`mana-segment flex-1 h-full border-r border-gray-700 last:border-r-0 transition-all ${i < game.playerMana
                                    ? "bg-gradient-to-r from-blue-400 to-blue-400"
                                    : ""
                                    }`}
                            />
                        ))}
                    </div>
                </div>

                {/* TROOP SELECTION */}
                <div className="troops-container bg-gradient-to-r from-blue-900 to-blue-800 p-4 rounded-lg mt-2 shadow-md border-2 border-blue-700">
                    <div className="section-header flex justify-between items-center mb-3">
                        <h3 className="text-xl font-bold text-yellow-400 drop-shadow-md">
                            TROOPS
                        </h3>

                        <button
                            className={`skip-btn px-4 py-1 rounded-full font-semibold transition-all transform hover:scale-105 ${game.playerTurn === user.user?.username
                                ? "bg-yellow-400 text-blue-900 border-2 border-yellow-500"
                                : "bg-gray-500 text-white opacity-50 cursor-not-allowed"
                                }`}
                            disabled={game.playerTurn !== user.user?.username}
                            onClick={skipTurn}
                        >
                            Skip
                        </button>
                    </div>

                    <div className="troop-selection flex flex-wrap justify-center gap-3">
                        {Object.entries(game.troops).map(([troopName, troop], index) => (
                            <div
                                key={index}
                                className={`troop w-50 ${game.selectedTroop?.name === troopName
                                    ? "border-4 border-yellow-400 bg-yellow-100 transform scale-105"
                                    : "border-2 border-gray-400 bg-white"
                                    } ${game.playerMana < troop.mana
                                        ? "opacity-50 grayscale"
                                        : "hover:scale-105"
                                    } rounded-lg shadow-lg p-2 cursor-pointer transition-all duration-200`}
                                onClick={() => selectTroop(troopName)}
                            >
                                <div className="troop-banner bg-gradient-to-r from-blue-600 to-blue-500 rounded-t-md px-2 py-1 -mt-2 -mx-2 mb-1 text-center">
                                    <div className="troop-name text-white font-bold drop-shadow-md truncate">
                                        {troopName}
                                    </div>
                                </div>
                                <div className="flex justify-between items-center mb-1">
                                    <div className="troop-mana-cost bg-blue-500 text-white font-bold flex items-center rounded-full px-2 border-2 border-blue-600">
                                        <span className="text-yellow-300 mr-1">⚡</span>{" "}
                                        {troop.mana}
                                    </div>
                                    {troop.atk && (
                                        <div className="troop-damage bg-red-500 text-white font-bold flex items-center rounded-full px-2 border-2 border-red-600">
                                            <span className="text-yellow-300 mr-1">💥</span>{" "}
                                            {troop.atk}
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
                {/* NOTIFICATION */}
                {notification.show && (
                    <div className="notification fixed top-5 left-0 right-0 mx-auto max-w-md bg-yellow-400 px-4 py-2 text-center rounded-full shadow-lg border-2 border-yellow-500 animate-bounce">
                        <span className="font-bold text-blue-900">
                            {notification.message}
                        </span>
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

                            <h2
                                className={`modal-title text-3xl font-bold mb-4 text-center ${game.winner === user.user?.username
                                    ? "text-yellow-400"
                                    : "text-red-400"
                                    }`}
                            >
                                {game.winner === user.user?.username ? "VICTORY!" : "DEFEAT"}
                            </h2>

                            <div className="modal-body text-center">
                                <p className="text-white text-lg mb-3">
                                    {game.winner === user.user?.username
                                        ? "You have conquered your opponent's kingdom!"
                                        : "Your kingdom has fallen to the enemy!"}
                                </p>
                                <div className="exp-gain text-yellow-300 font-bold text-2xl mt-3 animate-pulse">
                                    {game.winner === user.user?.username ? "+50 XP" : "+5 XP"}
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
        </div>
    );
}