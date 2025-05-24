import { Crown, Diamond, Eye, Gem, Heart, Shield, Sparkles, Star, Swords, Zap } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useNavigate } from "react-router-dom";
import { useWebSocketContext } from "../context/WebSocketContext";

const CardDesk = () => {
    const navigate = useNavigate();
    const { sendMessage, subscribe } = useWebSocketContext();

    const [troops, setTroops] = useState([]);
    const [selectedTroop, setSelectedTroop] = useState(null);

    // Load troop data and assign rarity based on mana cost
    useEffect(() => {
        if (!localStorage.getItem("session_id")) {
            setTimeout(() => navigate("/auth"), 1500);
            return;
        }

        // Gửi request sau khi page đã mount
        sendMessage({ type: "get_desk" });

        const unsubscribe = subscribe((res) => {
            console.log("Received WS msg:", res);
            if (res.type === "deck_response" && res.success) {
                if (troops.length === 0) {
                    setTroops(res.data);
                    console.log("Troops set:", res.data);
                } else {
                    console.log("Troops already set, skipping update.");
                }
            }
        });

        return () => unsubscribe();
    }, [navigate, subscribe, sendMessage, troops.length]);

    const getRarityConfig = (rarity) => {
        switch (rarity) {
            case 'common':
                return {
                    color: 'from-gray-400 to-gray-500',
                    borderColor: 'border-gray-400',
                    bgColor: 'bg-gray-100',
                    textColor: 'text-gray-700',
                    icon: <span className="text-xl">⭐</span>
                };
            case 'rare':
                return {
                    color: 'from-orange-400 to-orange-500',
                    borderColor: 'border-orange-400',
                    bgColor: 'bg-orange-100',
                    textColor: 'text-orange-700',
                    icon: <span className="text-xl">✨</span>
                };
            case 'epic':
                return {
                    color: 'from-purple-400 to-purple-500',
                    borderColor: 'border-purple-400',
                    bgColor: 'bg-purple-100',
                    textColor: 'text-purple-700',
                    icon: <span className="text-xl">💎</span>
                };
            case 'legendary':
                return {
                    color: 'from-blue-400 to-blue-500',
                    borderColor: 'border-blue-400',
                    bgColor: 'bg-blue-100',
                    textColor: 'text-blue-700',
                    icon: <span className="text-xl">👑</span>
                };
            case 'champion':
                return {
                    color: 'from-red-600 to-red-700',
                    borderColor: 'border-red-600',
                    bgColor: 'bg-red-100',
                    textColor: 'text-red-700',
                    icon: <span className="text-xl">🔥</span>
                };
            default:
                return {
                    color: 'from-gray-400 to-gray-500',
                    borderColor: 'border-gray-400',
                    bgColor: 'bg-gray-100',
                    textColor: 'text-gray-700',
                    icon: <span className="text-xl">⭐</span>
                };
        }
    };

    const getTypeColor = (type) => {
        switch (type) {
            case 'damage dealer': return 'from-red-500 to-red-600';
            case 'healer': return 'from-green-500 to-green-600';
            case 'tank': return 'from-blue-500 to-blue-600';
            default: return 'from-gray-500 to-gray-600';
        }
    };

    const getTypeIcon = (type) => {
        switch (type) {
            case 'damage dealer': return <span className="text-[0.8rem]">⚔️</span>;
            case 'healer': return <span className="text-[0.8rem]">❤️</span>;
            case 'tank': return <span className="text-[0.8rem]">🛡️</span>;
            default: return <span className="text-[0.8rem]">👁️</span>;
        }
    };

    // Group troops by rarity
    const groupedTroops = {
        common: troops.filter(t => t.rarity === 'common'),
        rare: troops.filter(t => t.rarity === 'rare'),
        epic: troops.filter(t => t.rarity === 'epic'),
        legendary: troops.filter(t => t.rarity === 'legendary'),
        champion: troops.filter(t => t.rarity === 'champion')
    };

    return (
        <div className={`min-h-screen bg-gradient-to-br from-sky-500 to-blue-600 p-4 md:p-8 font-sans ${selectedTroop ? 'overflow-hidden' : 'overflow-auto'}`}
            style={{ fontFamily: "'ClashDisplay', sans-serif" }}>
            <div className="max-w-6xl mx-auto mb-6">
                {/* Header */}
                <div className="text-center mb-8">
                    <div className="relative inline-block">
                        <h1
                            className="text-5xl md:text-6xl font-black text-white mb-2 drop-shadow-lg pointer-events-none"
                            style={{ textShadow: "3px 3px 0 #2563eb, 6px 6px 0 #1d4ed8" }}
                        >
                            <span className="text-yellow-400">COLLE</span>
                            <span className="text-red-500">CTION</span>
                        </h1>
                        <img
                            className="absolute w-10 -top-4 -right-4 transform rotate-20 pointer-events-none drop-shadow-[0_0_10px_rgba(255,255,0,0.6)]"
                            src="/assets/icon_crown.png"
                            alt=""
                        />
                        <img
                            className="absolute w-12 -bottom-2 -left-6 transform -rotate-12 pointer-events-none drop-shadow-[0_0_10px_rgba(255,255,0,0.6)]"
                            src="/assets/icon_badge.png"
                            alt=""
                        />
                    </div>
                    <div className="bg-blue-900 inline-block px-6 py-2 rounded-xl border-4 border-yellow-400 shadow-lg transform -rotate-1 pointer-events-none">
                        <p className="text-lg text-yellow-300">
                            Troop Cards Deck
                        </p>
                    </div>
                </div>
            </div>

            <div className="max-w-6xl mx-auto grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Troop Collection by Rarity */}
                <div className="lg:col-span-3 space-y-6">
                    {/* Common Cards */}
                    {groupedTroops.common.length > 0 && (
                        <div className="bg-gradient-to-r from-gray-400 via-gray-500 via-slate-400 to-gray-600 p-2 rounded-2xl">
                            <div className="bg-white rounded-2xl shadow-2xl p-6">
                                <div className="flex items-center mb-4">
                                    {getRarityConfig('common').icon}
                                    <span className="capitalize bg-gradient-to-r from-gray-400 via-gray-500 via-slate-400 to-gray-600 bg-clip-text text-transparent text-xl ml-2">Common Cards ({groupedTroops.common.length})</span>
                                </div>
                                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
                                    {groupedTroops.common.map((troop) => (
                                        <div
                                            key={troop.name}
                                            className={`aspect-[3/4] rounded-lg border-2 cursor-pointer transform hover:scale-110 transition-all duration-200 ${selectedTroop?.name === troop.name
                                                ? 'border-yellow-400 ring-2 ring-yellow-300 shadow-xl'
                                                : 'border-gray-300 hover:border-gray-400'
                                                } bg-gradient-to-br ${getTypeColor(troop.type)} relative overflow-hidden shadow-md`}
                                            onClick={() => setSelectedTroop(troop)}
                                        >
                                            <img
                                                src={`assets/cards/Card_${troop.image}.png`}
                                                alt={troop.name}
                                                className="w-full h-full object-cover"
                                            />

                                            {/* Card Name */}
                                            <div className="absolute bottom-0 left-0 right-0 bg-black bg-opacity-80 text-white text-sm p-1 text-center">
                                                {troop.name}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Rare Cards */}
                    {groupedTroops.rare.length > 0 && (
                        <div className="bg-gradient-to-r from-blue-400 via-cyan-500 via-sky-400 to-blue-600 p-2 rounded-2xl">
                            <div className="bg-white rounded-2xl shadow-2xl p-6">
                                <div className="flex items-center mb-4">
                                    {getRarityConfig('rare').icon}
                                    <span className="capitalize bg-gradient-to-r from-blue-400 via-cyan-500 via-sky-400 to-blue-600 bg-clip-text text-transparent text-xl ml-2">Rare Cards ({groupedTroops.rare.length})</span>
                                </div>
                                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
                                    {groupedTroops.rare.map((troop) => (
                                        <div
                                            key={troop.name}
                                            className={`aspect-[3/4] rounded-lg border-2 cursor-pointer transform hover:scale-110 transition-all duration-200 ${selectedTroop?.name === troop.name
                                                ? 'border-yellow-400 ring-2 ring-yellow-300 shadow-xl'
                                                : 'border-gray-300 hover:border-gray-400'
                                                } bg-gradient-to-br ${getTypeColor(troop.type)} relative overflow-hidden shadow-md`}
                                            onClick={() => setSelectedTroop(troop)}
                                        >
                                            <img
                                                src={`assets/cards/Card_${troop.image}.png`}
                                                alt={troop.name}
                                                className="w-full h-full object-cover"
                                            />

                                            {/* Card Name */}
                                            <div className="absolute bottom-0 left-0 right-0 bg-black bg-opacity-80 text-white text-sm p-1 text-center">
                                                {troop.name}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Epic Cards */}
                    {groupedTroops.epic.length > 0 && (
                        <div className="bg-gradient-to-r from-purple-500 via-violet-600 via-fuchsia-500 to-pink-600 p-2 rounded-2xl">
                            <div className=" bg-white rounded-2xl shadow-2xl p-6">
                                <div className="flex items-center mb-4">
                                    {getRarityConfig('epic').icon}
                                    <span className="capitalize bg-gradient-to-r from-purple-500 via-violet-600 via-fuchsia-500 to-pink-600 bg-clip-text text-transparent text-xl ml-2">Epic Cards ({groupedTroops.epic.length})</span>
                                </div>
                                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
                                    {groupedTroops.epic.map((troop) => (
                                        <div
                                            key={troop.name}
                                            className={`aspect-[3/4] rounded-lg border-2 cursor-pointer transform hover:scale-110 transition-all duration-200 ${selectedTroop?.name === troop.name
                                                ? 'border-yellow-400 ring-2 ring-yellow-300 shadow-xl'
                                                : 'border-gray-300 hover:border-gray-400'
                                                } bg-gradient-to-br ${getTypeColor(troop.type)} relative overflow-hidden shadow-md`}
                                            onClick={() => setSelectedTroop(troop)}
                                        >
                                            <img
                                                src={`assets/cards/Card_${troop.image}.png`}
                                                alt={troop.name}
                                                className="w-full h-full object-cover"
                                            />

                                            {/* Card Name */}
                                            <div className="absolute bottom-0 left-0 right-0 bg-black bg-opacity-80 text-white text-sm p-1 text-center">
                                                {troop.name}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Legendary Cards */}
                    {groupedTroops.legendary.length > 0 && (
                        <div className="bg-gradient-to-r from-amber-400 via-orange-500 via-red-500 to-pink-600 p-2 rounded-2xl">
                            <div className="bg-white rounded-2xl shadow-2xl p-6">
                                <div className="flex items-center mb-4">
                                    {getRarityConfig('legendary').icon}
                                    <span className="capitalize bg-gradient-to-r from-amber-400 via-orange-500 via-red-500 to-pink-600 bg-clip-text text-transparent text-xl ml-2">Legendary Cards ({groupedTroops.legendary.length})</span>
                                </div>
                                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
                                    {groupedTroops.legendary.map((troop) => (
                                        <div
                                            key={troop.name}
                                            className={`aspect-[3/4] rounded-lg border-2 cursor-pointer transform hover:scale-110 transition-all duration-200 ${selectedTroop?.name === troop.name
                                                ? 'border-yellow-400 ring-2 ring-yellow-300 shadow-xl'
                                                : 'border-gray-300 hover:border-gray-400'
                                                } bg-gradient-to-br ${getTypeColor(troop.type)} relative overflow-hidden shadow-md`}
                                            onClick={() => setSelectedTroop(troop)}
                                        >
                                            <img
                                                src={`assets/cards/Card_${troop.image}.png`}
                                                alt={troop.name}
                                                className="w-full h-full object-cover"
                                            />

                                            {/* Card Name */}
                                            <div className="absolute bottom-0 left-0 right-0 bg-black bg-opacity-80 text-white text-sm p-1 text-center">
                                                {troop.name}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Champion Cards */}
                    {groupedTroops.legendary.length > 0 && (
                        <div className="bg-gradient-to-r bg-gradient-to-r from-cyan-400 via-blue-500 via-purple-600 to-pink-600 p-2 rounded-2xl">
                            <div className="bg-white rounded-2xl shadow-2xl p-6">
                                <div className="flex items-center mb-4">
                                    {getRarityConfig('champion').icon}
                                    <span className="capitalize bg-gradient-to-r from-cyan-400 via-blue-500 via-purple-600 to-pink-600 bg-clip-text text-transparent text-xl ml-2">Champion Cards ({groupedTroops.champion.length})</span>
                                </div>
                                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
                                    {groupedTroops.champion.map((troop) => (
                                        <div
                                            key={troop.name}
                                            className={`aspect-[3/4] rounded-lg border-2 cursor-pointer transform hover:scale-110 transition-all duration-200 ${selectedTroop?.name === troop.name
                                                ? 'border-yellow-400 ring-2 ring-yellow-300 shadow-xl'
                                                : 'border-gray-300 hover:border-gray-400'
                                                } bg-gradient-to-br ${getTypeColor(troop.type)} relative overflow-hidden shadow-md`}
                                            onClick={() => setSelectedTroop(troop)}
                                        >
                                            <img
                                                src={`assets/cards/Card_${troop.image}.png`}
                                                alt={troop.name}
                                                className="w-full h-full object-cover"
                                            />

                                            {/* Card Name */}
                                            <div className="absolute bottom-0 left-0 right-0 bg-black bg-opacity-80 text-white text-sm p-1 text-center">
                                                {troop.name}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                {/* Selected Troop Details */}
                {selectedTroop && (
    <div
        className="fixed bg-black/50 backdrop-blur-sm inset-0 flex items-center justify-center p-2 z-50 shadow-3xl shadow-black min-h-screen"
        onClick={() => setSelectedTroop(null)} // Close when clicking on backdrop
    >
        {/* Dialog Header */}
        <div className="absolute top-[15vh] sm:top-[5vh] max-w-xs sm:max-w-md w-full mx-2 sm:mx-0 rounded-t-3xl border-t-4 border-l-4 border-r-4 border-emerald-400 bg-gradient-to-r from-emerald-500 to-cyan-500 z-20 p-3 sm:p-4 shadow-lg">
            <h2 className="text-lg sm:text-xl text-white text-center sm:text-left">Card Details</h2>
        </div>
        <div
            className="bg-white rounded-3xl shadow-2xl max-w-xs sm:max-w-md w-full max-h-[70vh] sm:max-h-[90vh] overflow-y-auto overflow-x-hidden no-scrollbar border-4 border-emerald-400 animate-in zoom-in duration-300 mx-2 sm:mx-0"
            onClick={(e) => e.stopPropagation()} // Prevent closing when clicking inside dialog
        >
            <div className="p-4 sm:p-6 mt-12 sm:mt-15">
                {/* Card Preview */}
                <div className="text-center mb-4 sm:mb-6">
                    <div className="relative inline-block">
                        <div className={`w-40 h-52 sm:w-56 sm:h-72 mx-auto rounded-xl border-4 border-yellow-400 shadow-2xl bg-gradient-to-br ${getTypeColor(selectedTroop.type)} overflow-hidden`}>
                            <div className="w-full h-full bg-gradient-to-br from-slate-100 to-slate-200 flex items-center justify-center">
                                <img
                                    src={`assets/cards/Card_${selectedTroop.image}.png`}
                                    alt={selectedTroop.name}
                                    className="w-full h-full object-cover"
                                />
                            </div>

                            <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black via-black/70 to-transparent text-white text-sm sm:text-lg p-2 sm:p-4 text-center rounded-b-xl">
                                {selectedTroop.name}
                            </div>
                        </div>
                    </div>
                </div>

                {/* Rarity Badge */}
                <div className="text-center mb-4 sm:mb-6">
                    <div className={`inline-flex items-center gap-2 px-4 py-2 sm:px-6 sm:py-3 rounded-full ${getRarityConfig(selectedTroop.rarity).bgColor} ${getRarityConfig(selectedTroop.rarity).textColor} text-base sm:text-lg shadow-lg border-2 ${getRarityConfig(selectedTroop.rarity).borderColor}`}>
                        <span className="text-lg sm:text-2xl">{getRarityConfig(selectedTroop.rarity).icon}</span>
                        <span className="capitalize">{selectedTroop.rarity}</span>
                    </div>
                </div>

                {/* Main Stats */}
                <div className="grid grid-cols-2 gap-2 sm:gap-4 mb-4 sm:mb-6">
                    <div className="bg-gradient-to-br from-red-50 to-red-100 p-3 sm:p-4 rounded-xl border-2 border-red-200 shadow-sm">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-1 sm:gap-2">
                                <span className="text-red-500 text-lg sm:text-xl">❤️</span>
                                <span className="text-red-600 font-semibold text-sm sm:text-base">HP</span>
                            </div>
                            <span className="text-red-700 text-lg sm:text-xl font-bold">{selectedTroop.hp}</span>
                        </div>
                    </div>
                    <div className="bg-gradient-to-br from-orange-50 to-orange-100 p-3 sm:p-4 rounded-xl border-2 border-orange-200 shadow-sm">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-1 sm:gap-2">
                                <span className="text-orange-500 text-lg sm:text-xl">⚔️</span>
                                <span className="text-orange-600 font-semibold text-sm sm:text-base">ATK</span>
                            </div>
                            <span className="text-orange-700 text-lg sm:text-xl font-bold">{selectedTroop.atk}</span>
                        </div>
                    </div>
                    <div className="bg-gradient-to-br from-blue-50 to-blue-100 p-3 sm:p-4 rounded-xl border-2 border-blue-200 shadow-sm">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-1 sm:gap-2">
                                <span className="text-blue-500 text-lg sm:text-xl">🛡️</span>
                                <span className="text-blue-600 font-semibold text-sm sm:text-base">DEF</span>
                            </div>
                            <span className="text-blue-700 text-lg sm:text-xl font-bold">{selectedTroop.def}</span>
                        </div>
                    </div>
                    <div className="bg-gradient-to-br from-purple-50 to-purple-100 p-3 sm:p-4 rounded-xl border-2 border-purple-200 shadow-sm">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-1 sm:gap-2">
                                <span className="text-purple-500 text-lg sm:text-xl">⚡</span>
                                <span className="text-purple-600 font-semibold text-sm sm:text-base">COST</span>
                            </div>
                            <span className="text-purple-700 text-lg sm:text-xl font-bold">{selectedTroop.mana}</span>
                        </div>
                    </div>
                </div>

                {/* Secondary Stats */}
                <div className="grid grid-cols-2 gap-2 sm:gap-3 mb-4 sm:mb-6">
                    <div className="bg-gradient-to-br from-green-50 to-green-100 p-2 sm:p-3 rounded-xl border border-green-200">
                        <div className="flex items-center justify-between">
                            <span className="text-green-600 font-medium text-xs sm:text-sm">🎯 CRIT</span>
                            <span className="text-green-700 text-sm sm:text-base font-semibold">{selectedTroop.crit}%</span>
                        </div>
                    </div>
                    <div className="bg-gradient-to-br from-yellow-50 to-yellow-100 p-2 sm:p-3 rounded-xl border border-yellow-200">
                        <div className="flex items-center justify-between">
                            <span className="text-yellow-600 font-medium text-xs sm:text-sm">🔥 EXP</span>
                            <span className="text-yellow-700 text-sm sm:text-base font-semibold">{selectedTroop.exp}</span>
                        </div>
                    </div>
                    <div className="bg-gradient-to-br from-indigo-50 to-indigo-100 p-2 sm:p-3 rounded-xl border border-indigo-200">
                        <div className="flex items-center justify-between">
                            <span className="text-indigo-600 font-medium text-xs sm:text-sm">⏱️ SPD</span>
                            <span className="text-indigo-700 text-sm sm:text-base font-semibold">{selectedTroop.attack_speed}</span>
                        </div>
                    </div>
                    <div className="bg-gradient-to-br from-pink-50 to-pink-100 p-2 sm:p-3 rounded-xl border border-pink-200">
                        <div className="flex items-center justify-between">
                            <span className="text-pink-600 font-medium text-xs sm:text-sm">📡 RNG</span>
                            <span className="text-pink-700 text-sm sm:text-base font-semibold">{selectedTroop.range}</span>
                        </div>
                    </div>
                </div>

                {/* Type Badge */}
                <div className="text-center mb-4 sm:mb-6">
                    <div className={`inline-flex items-center gap-2 sm:gap-3 px-4 py-2 sm:px-6 sm:py-3 rounded-full text-white font-semibold bg-gradient-to-r ${getTypeColor(selectedTroop.type)} shadow-lg`}>
                        <span className="text-lg sm:text-xl">{getTypeIcon(selectedTroop.type)}</span>
                        <span className="capitalize text-base sm:text-lg">{selectedTroop.type}</span>
                    </div>
                </div>

                {/* Description */}
                <div className="bg-gradient-to-br from-gray-50 to-gray-100 p-4 sm:p-5 rounded-xl border-2 border-gray-200 shadow-inner">
                    <p className="text-gray-700 leading-relaxed text-center text-sm sm:text-lg">
                        "{selectedTroop.description}"
                    </p>
                </div>
            </div>
        </div>
    </div>
)}
            </div>
        </div>
    );
};

export default CardDesk;