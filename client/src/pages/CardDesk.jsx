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
                    icon: <span className="text-sm">⭐</span>
                };
            case 'rare':
                return {
                    color: 'from-orange-400 to-orange-500',
                    borderColor: 'border-orange-400',
                    bgColor: 'bg-orange-100',
                    textColor: 'text-orange-700',
                    icon: <span className="text-sm">✨</span>
                };
            case 'epic':
                return {
                    color: 'from-purple-400 to-purple-500',
                    borderColor: 'border-purple-400',
                    bgColor: 'bg-purple-100',
                    textColor: 'text-purple-700',
                    icon: <span className="text-sm">💎</span>
                };
            case 'legendary':
                return {
                    color: 'from-blue-400 to-blue-500',
                    borderColor: 'border-blue-400',
                    bgColor: 'bg-blue-100',
                    textColor: 'text-blue-700',
                    icon: <span className="text-sm">👑</span>
                };
            case 'champion':
                return {
                    color: 'from-red-600 to-red-700',
                    borderColor: 'border-red-600',
                    bgColor: 'bg-red-100',
                    textColor: 'text-red-700',
                    icon: <span className="text-sm">🔥</span>
                };
            default:
                return {
                    color: 'from-gray-400 to-gray-500',
                    borderColor: 'border-gray-400',
                    bgColor: 'bg-gray-100',
                    textColor: 'text-gray-700',
                    icon: <span className="text-sm">⭐</span>
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
            case 'damage dealer': return <span className="text-xs">⚔️</span>;
            case 'healer': return <span className="text-xs">❤️</span>;
            case 'tank': return <span className="text-xs">🛡️</span>;
            default: return <span className="text-xs">👁️</span>;
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
        <div className="min-h-screen bg-gradient-to-br from-sky-500 to-blue-600 p-4 md:p-8 font-sans"
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

            <div className="max-w-6xl mx-auto grid grid-cols-1 lg:grid-cols-4 gap-6">
                {/* Troop Collection by Rarity */}
                <div className="lg:col-span-3 space-y-6">
                    {/* Common Cards */}
                    {groupedTroops.common.length > 0 && (
                        <div className="bg-white rounded-2xl shadow-2xl p-6 border-4 border-gray-400">
                            <div className="flex items-center mb-4">
                                {getRarityConfig('common').icon}
                                <h2 className="text-xl font-bold text-gray-700 ml-2">Common Cards ({groupedTroops.common.length})</h2>
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

                                        {/* Mana Cost */}
                                        <div className="absolute top-1 left-1 bg-purple-600 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs font-bold shadow-md">
                                            {troop.mana}
                                        </div>

                                        {/* Type Icon */}
                                        <div className="absolute top-1 right-1 bg-black bg-opacity-70 text-white rounded-full w-5 h-5 flex items-center justify-center shadow-md">
                                            {getTypeIcon(troop.type)}
                                        </div>

                                        {/* Card Name */}
                                        <div className="absolute bottom-0 left-0 right-0 bg-black bg-opacity-80 text-white text-sm p-1 text-center">
                                            {troop.name}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Rare Cards */}
                    {groupedTroops.rare.length > 0 && (
                        <div className="bg-white rounded-2xl shadow-2xl p-6 border-4 border-orange-400">
                            <div className="flex items-center mb-4">
                                {getRarityConfig('rare').icon}
                                <h2 className="text-xl font-bold text-orange-700 ml-2">Rare Cards ({groupedTroops.rare.length})</h2>
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

                                        {/* Mana Cost */}
                                        <div className="absolute top-1 left-1 bg-purple-600 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs font-bold shadow-md">
                                            {troop.mana}
                                        </div>

                                        {/* Type Icon */}
                                        <div className="absolute top-1 right-1 bg-black bg-opacity-70 text-white rounded-full w-5 h-5 flex items-center justify-center shadow-md">
                                            {getTypeIcon(troop.type)}
                                        </div>

                                        {/* Card Name */}
                                        <div className="absolute bottom-0 left-0 right-0 bg-black bg-opacity-80 text-white text-sm p-1 text-center">
                                            {troop.name}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Epic Cards */}
                    {groupedTroops.epic.length > 0 && (
                        <div className="bg-white rounded-2xl shadow-2xl p-6 border-4 border-purple-400">
                            <div className="flex items-center mb-4">
                                {getRarityConfig('epic').icon}
                                <h2 className="text-xl font-bold text-purple-700 ml-2">Epic Cards ({groupedTroops.epic.length})</h2>
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

                                        {/* Mana Cost */}
                                        <div className="absolute top-1 left-1 bg-purple-600 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs font-bold shadow-md">
                                            {troop.mana}
                                        </div>

                                        {/* Type Icon */}
                                        <div className="absolute top-1 right-1 bg-black bg-opacity-70 text-white rounded-full w-5 h-5 flex items-center justify-center shadow-md">
                                            {getTypeIcon(troop.type)}
                                        </div>

                                        {/* Card Name */}
                                        <div className="absolute bottom-0 left-0 right-0 bg-black bg-opacity-80 text-white text-sm p-1 text-center">
                                            {troop.name}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Legendary Cards */}
                    {groupedTroops.legendary.length > 0 && (
                        <div className="bg-white rounded-2xl shadow-2xl p-6 border-4 border-blue-400">
                            <div className="flex items-center mb-4">
                                {getRarityConfig('legendary').icon}
                                <h2 className="text-xl font-bold text-sky-700 ml-2">Legendary Cards ({groupedTroops.legendary.length})</h2>
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

                                        {/* Mana Cost */}
                                        <div className="absolute top-1 left-1 bg-purple-600 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs font-bold shadow-md">
                                            {troop.mana}
                                        </div>

                                        {/* Type Icon */}
                                        <div className="absolute top-1 right-1 bg-black bg-opacity-70 text-white rounded-full w-5 h-5 flex items-center justify-center shadow-md">
                                            {getTypeIcon(troop.type)}
                                        </div>

                                        {/* Card Name */}
                                        <div className="absolute bottom-0 left-0 right-0 bg-black bg-opacity-80 text-white text-sm p-1 text-center">
                                            {troop.name}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Champion Cards */}
                    {groupedTroops.legendary.length > 0 && (
                        <div className="bg-white rounded-2xl shadow-2xl p-6 border-4 border-yellow-400">
                            <div className="flex items-center mb-4">
                                {getRarityConfig('legendary').icon}
                                <h2 className="text-xl font-bold text-yellow-700 ml-2">Champion Cards ({groupedTroops.champion.length})</h2>
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

                                        {/* Mana Cost */}
                                        <div className="absolute top-1 left-1 bg-purple-600 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs font-bold shadow-md">
                                            {troop.mana}
                                        </div>

                                        {/* Type Icon */}
                                        <div className="absolute top-1 right-1 bg-black bg-opacity-70 text-white rounded-full w-5 h-5 flex items-center justify-center shadow-md">
                                            {getTypeIcon(troop.type)}
                                        </div>

                                        {/* Card Name */}
                                        <div className="absolute bottom-0 left-0 right-0 bg-black bg-opacity-80 text-white text-sm p-1 text-center">
                                            {troop.name}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}
                </div>

                {/* Selected Troop Details */}
                <div className="lg:col-span-1">
                    {selectedTroop && (
                        <div className="bg-white rounded-2xl shadow-2xl p-6 border-4 border-green-400 sticky top-4">
                            <div className="text-center mb-4">
                                <img
                                    src={`assets/cards/Card_${selectedTroop.image}.png`}
                                    alt={selectedTroop.name}
                                    className="w-40 h-50 mx-auto rounded-lg border-4 border-yellow-400 shadow-lg object-cover"
                                />
                            </div>

                            <h3 className="text-xl font-bold text-gray-800 text-center mb-2">{selectedTroop.name}</h3>

                            <div className={`text-center mb-4 px-3 py-1 rounded-full ${getRarityConfig(selectedTroop.rarity).bgColor} ${getRarityConfig(selectedTroop.rarity).textColor} font-semibold capitalize`}>
                                {selectedTroop.rarity}
                            </div>

                            <div className="grid grid-cols-2 gap-2 text-xs mb-3">
                                <div className="flex items-center justify-between bg-red-50 p-2 rounded-md">
                                    <div className="flex items-center">
                                        <span className="text-red-500 mr-1 text-sm">❤️</span>
                                        <span className="text-red-600">HP</span>
                                    </div>
                                    <span className="font-bold text-red-700">{selectedTroop.hp}</span>
                                </div>
                                <div className="flex items-center justify-between bg-orange-50 p-2 rounded-md">
                                    <div className="flex items-center">
                                        <span className="text-orange-500 mr-1 text-sm">⚔️</span>
                                        <span className="text-orange-600">ATK</span>
                                    </div>
                                    <span className="font-bold text-orange-700">{selectedTroop.atk}</span>
                                </div>
                                <div className="flex items-center justify-between bg-blue-50 p-2 rounded-md">
                                    <div className="flex items-center">
                                        <span className="text-blue-500 mr-1 text-sm">🛡️</span>
                                        <span className="text-blue-600">DEF</span>
                                    </div>
                                    <span className="font-bold text-blue-700">{selectedTroop.def}</span>
                                </div>
                                <div className="flex items-center justify-between bg-purple-50 p-2 rounded-md">
                                    <div className="flex items-center">
                                        <span className="text-purple-500 mr-1 text-sm">⚡</span>
                                        <span className="text-purple-600">COST</span>
                                    </div>
                                    <span className="font-bold text-purple-700">{selectedTroop.mana}</span>
                                </div>
                            </div>

                            {/* Additional Stats */}
                            <div className="grid grid-cols-2 gap-2 text-xs mb-4">
                                <div className="flex items-center justify-between bg-green-50 p-2 rounded-md">
                                    <div className="flex items-center">
                                        <span className="text-green-500 mr-1 text-sm">🎯</span>
                                        <span className="text-green-600">CRIT</span>
                                    </div>
                                    <span className="font-bold text-green-700">{selectedTroop.crit}%</span>
                                </div>
                                <div className="flex items-center justify-between bg-yellow-50 p-2 rounded-md">
                                    <div className="flex items-center">
                                        <span className="text-yellow-500 mr-1 text-sm">🔥</span>
                                        <span className="text-yellow-600">EXP</span>
                                    </div>
                                    <span className="font-bold text-yellow-700">{selectedTroop.exp}</span>
                                </div>
                                <div className="flex items-center justify-between bg-indigo-50 p-2 rounded-md">
                                    <div className="flex items-center">
                                        <span className="text-gray-500 mr-1 text-sm">⏱️</span>
                                        <span className="text-indigo-600">SPD</span>
                                    </div>
                                    <span className="font-bold text-indigo-700">{selectedTroop.attack_speed}</span>
                                </div>
                                <div className="flex items-center justify-between bg-pink-50 p-2 rounded-md">
                                    <div className="flex items-center">
                                        <span className="text-pink-500 mr-1 text-sm">📡</span>
                                        <span className="text-pink-600">RNG</span>
                                    </div>
                                    <span className="font-bold text-pink-700">{selectedTroop.range}</span>
                                </div>
                            </div>

                            <div className="text-center mb-3">
                                <div className={`inline-flex items-center px-3 py-1 rounded-full text-white text-sm font-semibold bg-gradient-to-r ${getTypeColor(selectedTroop.type)}`}>
                                    {getTypeIcon(selectedTroop.type)}
                                    <span className="ml-1 capitalize">{selectedTroop.type}</span>
                                </div>
                            </div>

                            <p className="text-sm text-gray-600 text-center leading-relaxed">{selectedTroop.description}</p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default CardDesk;