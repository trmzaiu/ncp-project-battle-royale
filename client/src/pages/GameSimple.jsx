import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useWebSocketContext } from "../context/WebSocketContext";

export default function GameSimple() {
    const navigate = useNavigate();
    const { sendMessage, subscribe } = useWebSocketContext();
    const damageTimeoutRef = useRef(null);
    const healTimeoutRef = useRef(null);
    const hasLeftGameRef = useRef(false);

    const [user, setUser] = useState({});
    const [opponent, setOpponent] = useState({});
    const [isGameInitialized, setIsGameInitialized] = useState(false);
    const [showLargeAnimation, setShowLargeAnimation] = useState(false);
    const [hoveredTroop, setHoveredTroop] = useState(null);

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
            message: ""
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
            navigate("/auth")
            return;
        }

        // if (!localStorage.getItem("room_id")) {
        //     showNotification("Room not found. Redirecting to lobby...");
        //     navigate("/lobby")
        //     return;
        // }

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
            leaveGame();
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
            winner: "",
            message: ""
        });

        setIsGameInitialized(true);

        

        setShowLargeAnimation(true);
        setTimeout(() => {
            setShowLargeAnimation(false);
        }, 2000);
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

    // === Leave Game ===
    const leaveGame = () => {
        if (hasLeftGameRef.current) return;

        hasLeftGameRef.current = true;
        sendMessage({
            type: "leave_game",
            data: {
                room_id: localStorage.getItem("room_id"),
                username: localStorage.getItem("username"),
            },
        });
        localStorage.removeItem("room_id");
    };

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

        if (troop.type === "healer") {
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
        const { attacker, defender, damage, target, isDestroyed, turn, isCrit } =
            msg.data;
        const isMe = attacker.user.username === localStorage.getItem("username");
        const targetId = (isMe ? "opponent-" : "player-") + target;

        if (damageTimeoutRef.current) {
            clearTimeout(damageTimeoutRef.current);
            damageTimeoutRef.current = null;
            // Hide the current popup
            setDamagePopup(prev => ({ ...prev, visible: false }));
        }

        setGame((prev) => {
            const newState = { ...prev };

            if (isMe) {
                setUser(attacker);
                setOpponent(defender);
                newState.playerMana = attacker.mana;
                newState.opponentHealth[target] = defender.towers[target].hp;
            } else {
                setUser(defender);
                setOpponent(attacker);
                newState.playerMana = defender.mana;
                newState.playerHealth[target] = defender.towers[target].hp;
            }

            return newState;
        });

        setTimeout(() => {
            setDamagePopup({
                targetId,
                amount: damage,
                isOpponent: isMe,
                visible: true,
                crit: isCrit,
            });

            // Hide damage popup after 1.5 seconds
            damageTimeoutRef.current = setTimeout(() => {
                setDamagePopup(prev => ({ ...prev, visible: false }));
                damageTimeoutRef.current = null;

                // AFTER damage popup is hidden, update turn and show turn animation
                setTimeout(() => {
                    // Update the turn
                    setGame(prev => ({
                        ...prev,
                        playerTurn: turn
                    }));

                    // Only show turn animation if it's the player's turn now
                    setShowLargeAnimation(true);

                    // Hide turn animation after 2 seconds
                    setTimeout(() => {
                        setShowLargeAnimation(false);
                    }, 2000);
                }, 300); // Small delay after damage popup disappears

            }, 1500);
        }, 50);

        if (isDestroyed) showNotification(`Tower ${target} has been destroyed!`);
    };

    // === Handle Heal Response ===
    const handleHeal = (msg) => {
        const { turn, player, opponent, healedTower, healAmount } = msg.data;
        const isMe = player.user.username === localStorage.getItem("username");
        const targetId = (isMe ? "player-" : "opponent-") + healedTower.type;

        if (healTimeoutRef.current) {
            clearTimeout(healTimeoutRef.current);
            healTimeoutRef.current = null;
            // Hide the current popup
            setHealPopup(prev => ({ ...prev, visible: false }));
        }


        setGame((prev) => {
            const newState = { ...prev };

            if (isMe) {
                setUser(player);
                setOpponent(opponent);
                newState.playerMana = player.mana;
                newState.playerHealth[healedTower.type] = healedTower.hp;
            } else {
                setUser(opponent);
                setOpponent(player);
                newState.playerMana = opponent.mana;
                newState.opponentHealth[healedTower.type] = healedTower.hp;
            }

            return newState;

        });

        setTimeout(() => {
            setHealPopup({
                targetId,
                amount: healAmount,
                isOpponent: !isMe,
                visible: true,
            });

            // Hide damage popup after 1.5 seconds
            healTimeoutRef.current = setTimeout(() => {
                setHealPopup(prev => ({ ...prev, visible: false }));
                healTimeoutRef.current = null;

                // AFTER damage popup is hidden, update turn and show turn animation
                setTimeout(() => {
                    // Update the turn
                    setGame(prev => ({
                        ...prev,
                        playerTurn: turn
                    }));

                    // Only show turn animation if it's the player's turn now
                    setShowLargeAnimation(true);

                    // Hide turn animation after 2 seconds
                    setTimeout(() => {
                        setShowLargeAnimation(false);
                    }, 2000);
                }, 300); // Small delay after damage popup disappears

            }, 1500);
        }, 50);
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
            selectedTroop: null
        }));

        setShowLargeAnimation(true);

        setTimeout(() => {
            setShowLargeAnimation(false);
        }, 2000);
    };


    // === Handle Game Over ===
    const handleGameOver = (res) => {

        setShowLargeAnimation(false);
        setTimeout(() => {
            setGame((prev) => ({ ...prev, gameOver: true, winner: res.data.winner.user.username, message: res.message }));
        }, 2000)
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
        navigate("/lobby");
    };

    // Tower component for reusability 
    const Tower = ({ id, type, health, maxHealth, isOpponent, onClick, disabled }) => {
        const towerImage = type === "king"
            ? (isOpponent ? "/royaka-2025-fe/assets/King_Tower_Red.png" : "/royaka-2025-fe/assets/King_Tower_Blue.png")
            : (isOpponent ? "/royaka-2025-fe/assets/Guard_Tower_Red.png" : "/royaka-2025-fe/assets/Guard_Tower_Blue.png");

        return (
            <div
                className={`relative tower ${type} ${health <= 0 ? "grayscale" : ""} w-full h-full flex flex-col items-center justify-end ${!disabled ? "cursor-pointer" : ""}`}
                onClick={disabled ? undefined : onClick}
                style={{ fontFamily: "'ClashDisplay', sans-serif" }}
            >
                {/* Damage Popup */}
                {damagePopup?.visible && damagePopup.targetId === id && (
                    <div className={`absolute z-50 ${isOpponent ? "-bottom-15" : "-top-15"} -right-4 transform rotate-10 pointer-events-none`}>
                        <div className={`
                        flex justify-center items-center
                        ${damagePopup.crit ? "text-yellow-500 animate-clash-crit-popup" : "text-red-500 animate-clash-damage-popup"}
                    `}>
                            <div className="relative">
                                {["-left-1", "-right-1", "-top-1", "-bottom-1"].map((pos, i) => (
                                    <span key={i} className={`absolute font-extrabold ${damagePopup.crit ? "text-4xl" : "text-3xl"} text-white opacity-25 ${pos}`}>
                                        -{damagePopup.amount}
                                    </span>
                                ))}
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

                {/* Heal Popup */}
                {healPopup?.visible && healPopup.targetId === id && (
                    <div className={`absolute z-50 ${isOpponent ? "-bottom-15" : "-top-15"} rotate-10 pointer-events-none`}>
                        <div className="flex justify-center items-center text-green-700 animate-clash-heal-popup">
                            <div className="relative">
                                {["-left-0.5", "-right-0.5", "-top-0.5", "-bottom-0.5"].map((pos, i) => (
                                    <span key={i} className={`absolute font-extrabold text-3xl text-white opacity-25 ${pos}`}>
                                        +{healPopup.amount}
                                    </span>
                                ))}
                                <span className="relative font-extrabold text-3xl">
                                    +{healPopup.amount}
                                </span>
                            </div>
                            <div className="absolute w-12 h-12 bg-green-400 rounded-full opacity-20 animate-clash-heal-burst" />
                        </div>
                    </div>
                )}

                <div className="w-full h-full flex flex-col items-center justify-end relative">
                    {isOpponent ? (
                        <>
                            <img
                                src={towerImage}
                                alt={type}
                                className="relative -bottom-3 w-full h-full object-contain px-1 py-1 drop-shadow-lg transition-transform duration-200 hover:scale-105"
                            />
                            <div className="relative -bottom-5 tower-hp w-5/6">
                                <div className="hp-bar bg-gray-700 w-full h-3 rounded-full shadow-inner overflow-hidden border border-gray-800">
                                    <div
                                        className={`hp-fill bg-gradient-to-r ${health <= maxHealth / 3
                                            ? "from-red-500 to-red-400"
                                            : "from-green-500 to-green-400"
                                            } h-full rounded-full transition-all duration-500`}
                                        style={{
                                            width: `${Math.max(0, (health / maxHealth) * 100)}%`,
                                        }}
                                    />
                                </div>
                            </div>
                        </>
                    ) : (
                        <>
                            <div className="relative -top-3 tower-hp w-5/6">
                                <div className="hp-bar bg-gray-700 w-full h-3 rounded-full shadow-inner overflow-hidden border border-gray-800">
                                    <div
                                        className={`hp-fill bg-gradient-to-r ${health <= maxHealth / 3
                                            ? "from-red-500 to-red-400"
                                            : "from-green-500 to-green-400"
                                            } h-full rounded-full transition-all duration-500`}
                                        style={{
                                            width: `${Math.max(0, (health / maxHealth) * 100)}%`,
                                        }}
                                    />
                                </div>
                            </div>
                            <img
                                src={towerImage}
                                alt={type}
                                className="relative w-full h-full object-contain px-1 py-1 drop-shadow-lg transition-transform duration-200 hover:scale-105"
                            />
                        </>
                    )}
                </div>
            </div>
        );
    };

    const tileMap = [
        ["00", "01", "00", "87", "80", "80", "88", "00", "02", "02"],
        ["01", "00", "00", "87", "80", "95", "86", "00", "02", "01"],
        ["19", "20", "00", "85", "84", "86", "01", "00", "00", "00"],
        ["37", "72", "19", "19", "19", "19", "19", "19", "19", "19"],
        ["55", "55", "55", "55", "55", "55", "55", "55", "55", "96"],
        ["00", "00", "01", "00", "00", "01", "00", "00", "00", "36"],
        ["01", "00", "00", "00", "00", "00", "00", "00", "00", "36"],
        ["82", "83", "00", "01", "00", "00", "01", "00", "02", "54"]
    ];

    return (
        <div className="min-h-screen bg-gradient-to-b from-blue-900 via-blue-500 to-blue-900">
            <div
                className="game-container bg-gradient-to-b from-blue-400 to-blue-600 p-2 rounded-lg shadow-xl max-w-2xl mx-auto relative overflow-hidden border-4 border-yellow-500"
                style={{ fontFamily: "'ClashDisplay', sans-serif" }}
            >
                {/* Decorative elements */}
                <div className="absolute -top-16 -left-16 w-32 h-32 bg-yellow-300 rounded-full opacity-20"></div>
                <div className="absolute -bottom-16 -right-16 w-32 h-32 bg-yellow-300 rounded-full opacity-20"></div>

                {/* Game title */}
                <div className="text-center mb-2">
                    <h1 className="text-3xl text-yellow-400 drop-shadow-md transform rotate-2 mb-1">
                        ROYAKA
                    </h1>
                    <div className="w-32 h-1 bg-yellow-400 mx-auto rounded-full mb-2"></div>
                </div>

                {/* STATS BAR */}
                <div className="stats-bar flex justify-between items-center p-2 bg-gradient-to-r from-blue-900 to-blue-800 rounded-lg shadow-md mb-2">
                    {/* OPPONENT STATS */}
                    <div className="opponent-stats flex items-center">
                        <div className="opponent-avatar relative mx-2">
                            <img
                                src={`/royaka-2025-fe/assets/avatars/Avatar${opponent.user?.avatar}.png`}
                                alt="avatar"
                                className="w-12 h-12 rounded-full border-2 border-red-700 shadow-md transform hover:scale-105 transition-transform object-cover"
                            />
                            <div className="absolute -bottom-1 -right-1 bg-red-700 text-white text-xs rounded-full w-6 h-6 flex items-center justify-center border border-yellow-400">
                                {opponent.user?.level || 0}
                            </div>
                        </div>
                        <div className="stat-column ml-2">
                            <div className="stat">
                                <div className="stat-value name text-yellow-500 text-lg drop-shadow-md">
                                    {opponent.user?.username || "Waiting..."}
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* TURN DISPLAY */}
                    <div className="turn-display text-center transform hover:scale-105 transition-transform mx-2">
                        <div
                            className={`text-sm pt-2 pb-1 md:text-lg px-4 rounded-full ${game.playerTurn === user.user?.username
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
                <div className="battle-container rounded-lg shadow-inner border-4 border-green-700 relative w-full aspect-[10/8]">
                    {game.playerHealth['king'] > 0 && game.opponentHealth['king'] > 0 &&
                        showLargeAnimation && (
                            <div
                                className="absolute flex items-center justify-center z-50 pointer-events-none w-full h-full"
                                style={{
                                    fontFamily: "'ClashDisplay', sans-serif",
                                    textShadow: "2px 2px 10px rgba(0, 0, 0, 0.5)",
                                }}
                            >
                                <div className="text-2xl md:text-5xl text-white animate-turnAlert">
                                    {game.playerTurn === localStorage.getItem("username")
                                        ? "YOUR TURN"
                                        : "OPPONENT'S TURN"}
                                </div>
                            </div>
                        )
                    }

                    {game.selectedTroop && (
                        <div className="target-indicators absolute flex items-center justify-center z-50 pointer-events-none w-full h-full">
                            <div className="text-sm md:text-3xl text-white animate-turnAlert bg-blue-800 bg-opacity-70 px-4 py-2 rounded-full">
                                🎯 Pick a Target!
                            </div>
                        </div>
                    )}

                    {/* Grid background */}
                    <div className="absolute inset-0 grid grid-cols-10 grid-rows-8">
                        {tileMap.map((row, rowIndex) =>
                            row.map((tile, colIndex) => (
                                <div
                                    key={`r${rowIndex}-c${colIndex}`}
                                    className="bg-cover flex items-center justify-center"
                                    style={{
                                        backgroundImage: `url(/royaka-2025-fe/assets/tiles/tile_00${tile}.png)`,
                                    }}
                                >
                                    {colIndex === 0 && rowIndex === 0 && (
                                        <img
                                            src="/royaka-2025-fe/assets/tiles/tile_0092.png"
                                            alt="Tree"
                                            className="w-full h-full pointer-events-none select-none"
                                            style={{ objectFit: "contain" }}
                                        />
                                    )}
                                    {colIndex === 0 && rowIndex === 1 && (
                                        <img
                                            src="/royaka-2025-fe/assets/tiles/tile_0090.png"
                                            alt="Tree"
                                            className="w-full h-full pointer-events-none select-none"
                                            style={{ objectFit: "contain" }}
                                        />
                                    )}
                                    {colIndex === 1 && rowIndex === 0 && (
                                        <img
                                            src="/royaka-2025-fe/assets/tiles/tile_0090.png"
                                            alt="Tree"
                                            className="w-full h-full pointer-events-none select-none"
                                            style={{ objectFit: "contain" }}
                                        />
                                    )}
                                </div>
                            ))
                        )}
                    </div>

                    {/* Battlefield grid layout - 10 columns x 6 rows */}
                    <div className="grid-battlefield grid grid-cols-10 grid-rows-8 relative w-full aspect-[10/8]">
                        <div className="col-start-5 col-span-2 row-start-1 row-span-2 relative">
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

                        <div className="col-start-3 row-start-2 relative">
                            <Tower
                                type="guard"
                                id="opponent-guard1"
                                health={game.opponentHealth.guard1}
                                maxHealth={game.opponentShield.guard1}
                                isOpponent={true}
                                onClick={() => selectTarget("guard1")}
                                disabled={!game.selectedTroop}
                            />
                        </div>

                        <div className="col-start-8 row-start-2 relative">
                            <Tower
                                type="guard"
                                id="opponent-guard2"
                                health={game.opponentHealth.guard2}
                                maxHealth={game.opponentShield.guard2}
                                isOpponent={true}
                                onClick={() => selectTarget("guard2")}
                                disabled={!game.selectedTroop}
                            />
                        </div>

                        <div className="col-start-3 row-start-7 relative">
                            <Tower
                                type="guard"
                                id="player-guard1"
                                health={game.playerHealth.guard1}
                                maxHealth={game.playerShield.guard1}
                                isOpponent={false}
                                disabled={true}
                            />
                        </div>

                        <div className="col-start-8 row-start-7 relative">
                            <Tower
                                type="guard"
                                id="player-guard2"
                                health={game.playerHealth.guard2}
                                maxHealth={game.playerShield.guard2}
                                isOpponent={false}
                                disabled={true}
                            />
                        </div>

                        <div className="col-start-5 col-span-2 row-start-7 row-span-2 relative">
                            <Tower
                                type="king"
                                id="player-king"
                                health={game.playerHealth.king}
                                maxHealth={game.playerShield.king}
                                isOpponent={false}
                                disabled={true}
                                className="w-full h-full"
                            />
                        </div>
                    </div>

                    {/* Target indicators */}

                </div>

                {/* MANA BAR */}
                <div className="mana-container bg-gradient-to-r from-blue-900 to-blue-800 p-2 rounded-lg border-2 border-blue-700 my-2 shadow-md">
                    <div className="flex items-center justify-between mb-1">
                        <div className="text-lg text-yellow-400 flex items-center">
                            <span className="text-xl mr-1">⚡</span> MANA
                        </div>
                        <div className="text-xl text-white font-bold me-1">
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
                <div className="troops-container bg-gradient-to-r from-blue-900 to-blue-800 p-2 rounded-lg mt-2 shadow-md border-2 border-blue-700">
                    <div className="section-header flex justify-between items-center mb-2 px-1">
                        <h3 className="text-xl text-yellow-400 drop-shadow-md">TROOPS</h3>

                        <button
                            className={`skip-btn px-4 py-0.5 rounded-full font-semibold transition-all transform hover:scale-105 ${game.playerTurn === user.user?.username
                                ? "bg-yellow-400 text-blue-900"
                                : "bg-gray-500 text-white opacity-50 cursor-not-allowed"
                                }`}
                            disabled={game.playerTurn !== user.user?.username}
                            onClick={skipTurn}
                        >
                            Skip
                        </button>
                    </div>

                    <div className="troop-selection flex flex-wrap justify-between">
                        {Object.entries(game.troops).map(([troopName, troop], index) => (
                            <div key={index} className="relative"
                                onMouseEnter={() => setHoveredTroop(troopName)}
                                onMouseLeave={() => setHoveredTroop(null)}
                            >
                                {hoveredTroop === troopName && (
                                    <div className="absolute bottom-full mb-2 left-1/2 transform -translate-x-1/2 w-52 bg-gray-800 text-white rounded-lg shadow-lg z-20 overflow-hidden">
                                        {/* Troop Header */}
                                        <div className="bg-gradient-to-r from-blue-700 to-blue-900 p-2 border-b border-gray-600">
                                            <h4 className="text-yellow-300 text-center">{troopName}</h4>
                                        </div>

                                        {/* Troop Description */}
                                        <div className="p-2 pb-1 border-b border-gray-600 text-center italic text-[11px]">
                                            {troop.description}
                                        </div>

                                        {/* Troop Stats */}
                                        <div className="p-2">
                                            <div className="grid grid-cols-2 gap-x-3 gap-y-1 ">
                                                <div className="flex items-center">
                                                    <span className="w-5 h-5 rounded-md bg-red-200 flex items-center justify-center mr-1">❤️</span>
                                                    <span className="text-[11px]">HP: {troop.max_hp}</span>
                                                </div>
                                                <div className="flex items-center">
                                                    <span className="w-5 h-5 rounded-md bg-yellow-200 flex items-center justify-center mr-1">⚔️</span>
                                                    <span className="text-[11px]">ATK: {troop.atk}</span>
                                                </div>
                                                <div className="flex items-center">
                                                    <span className="w-5 h-5 rounded-md bg-green-200 flex items-center justify-center mr-1">🛡️</span>
                                                    <span className="text-[11px]">DEF: {troop.def}</span>
                                                </div>
                                                <div className="flex items-center">
                                                    <span className="w-5 h-5 rounded-md bg-purple-200 flex items-center justify-center mr-1">🎯</span>
                                                    <span className="text-[11px]">CRIT: {troop.crit}%</span>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                )}
                                <div
                                    className={`troop ${game.selectedTroop?.name === troopName
                                        ? "border-4 border-yellow-400 transform scale-105"
                                        : "border-2 border-gray-700"
                                        } ${game.playerMana < troop.mana
                                            ? "opacity-50 grayscale"
                                            : "hover:scale-105"
                                        } rounded-xl shadow-lg shadow-inner relative overflow-hidden cursor-pointer transition-all duration-200`}
                                    onClick={() => selectTroop(troopName)}
                                >
                                    <div className="w-full relative">
                                        <img
                                            className="w-20 h-22 md:w-37 md:h-37 object-cover"
                                            src={`/royaka-2025-fe/assets/cards/Card_${troop.image}.png`}
                                            alt={troopName}
                                        />

                                        {/* Mana cost */}
                                        <div className="absolute bottom-1 right-1 bg-blue-800 bg-opacity-80 rounded-full w-8 h-8 flex items-center justify-center shadow-md border border-blue-400">
                                            <span className="text-white text-lg mt-1 leading-none">{troop.mana}</span>
                                        </div>
                                    </div>
                                    <div className="bg-gradient-to-r from-gray-800 to-gray-900 p-1 text-center">
                                        <span className="text-white text-[9px] md:text-sm font-semibold truncate block">{troopName}</span>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>

                {/* NOTIFICATION */}
                {notification.show && (
                    <div className="notification fixed top-5 left-0 right-0 mx-auto max-w-md bg-yellow-400 px-4 py-2 text-center rounded-full shadow-lg border-2 border-yellow-500 animate-bounce">
                        <span className="text-blue-900">{notification.message}</span>
                    </div>
                )}

                {/* GAME OVER MODAL */}
                { game.gameOver && (
                    <div className="game-over-modal fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
                        <div className="modal-content bg-gradient-to-b from-blue-800 to-blue-900 rounded-lg shadow-xl p-6 max-w-md w-full border-4 border-yellow-500 transform scale-105">
                            {/* Crown decoration */}
                            <div className="crown-decoration absolute -top-10 left-1/2 transform -translate-x-1/2 text-6xl animate-pulse">
                                {game.winner === localStorage.getItem("username")
                                    ? "👑"
                                    : "☠️"}
                            </div>

                            <h2
                                className={`modal-title text-3xl mb-4 text-center animate-pulse ${game.winner === localStorage.getItem("username")
                                    ? "text-yellow-400"
                                    : "text-red-400"
                                    }`}
                            >
                                {game.winner === localStorage.getItem("username")
                                    ? "VICTORY!"
                                    : "DEFEAT"}
                            </h2>

                            <div className="modal-body text-center animate-pulse">
                                {(() => {
                                    let victoryMessage = "";
                                    if (game.winner === localStorage.getItem("username")) {
                                        victoryMessage = game.message !== ""
                                            ? "You have conquered your opponent's kingdom!"
                                            : "You won by opponent leaving the battle!";
                                    } else {
                                        victoryMessage = "Your kingdom has fallen to the enemy!";
                                    }
                                    return (
                                        <p className="text-white text-lg mb-3">
                                            {victoryMessage}
                                        </p>
                                    );
                                })()}
                                <div className="exp-gain text-yellow-300 text-2xl mt-3 animate-pulse">
                                    {game.winner === localStorage.getItem("username")
                                        ? "+30 XP"
                                        : ""}
                                </div>
                                <img src={game.winner === user.user?.username
                                    ? "/royaka-2025-fe/assets/win.png"
                                    : "/royaka-2025-fe/assets/lose.png"}
                                    alt={game.winner === user.user?.username
                                        ? "Winner"
                                        : "Loser"} className="w-40 h-40 mx-auto" />
                            </div>

                            <div className="modal-buttons text-center animate-pulse">
                                <button
                                    className="play-again-btn bg-gradient-to-r from-yellow-400 to-yellow-500 text-blue-900 px-8 py-3 rounded-full text-lg border-4 border-yellow-600 shadow-lg transform hover:scale-105 transition-transform"
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