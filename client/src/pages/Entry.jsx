import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";

const Entry = () => {
    const [cloudPosition, setCloudPosition] = useState(0);
    const [flagWave, setFlagWave] = useState(0);
    const navigate = useNavigate();

    // Animation timer for environmental elements
    useEffect(() => {
        const interval = setInterval(() => {
            setCloudPosition(pos => (pos + 1) % 100);
            setFlagWave(wave => (wave + 1) % 20);
        }, 100);

        return () => clearInterval(interval);
    }, []);

    // Navigation handling
    const handlePlay = () => {
        setTimeout(() => navigate("/auth"), 1500);
    };

    return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-gradient-to-b from-blue-400 to-blue-600 p-6 relative overflow-hidden">
            {/* Sky and clouds */}
            <div className="absolute top-0 left-0 w-full h-full">
                {/* Sun */}
                <div className="absolute top-16 right-16 w-20 h-20 bg-red-400 rounded-full"></div>

                {/* Clouds */}
                <div className="absolute top-24 w-full">
                    <div className={`absolute w-32 h-12 bg-white rounded-full opacity-80`}
                        style={{ left: `${(cloudPosition + 10) % 100}%`, transform: 'translateX(-50%)' }}></div>
                    <div className={`absolute w-40 h-16 bg-white rounded-full opacity-70 top-8`}
                        style={{ left: `${(cloudPosition + 40) % 100}%`, transform: 'translateX(-50%)' }}></div>
                    <div className={`absolute w-24 h-10 bg-white rounded-full opacity-75 top-16`}
                        style={{ left: `${(cloudPosition + 70) % 100}%`, transform: 'translateX(-50%)' }}></div>
                </div>
            </div>

            {/* Background hills */}
            <div className="absolute bottom-0 left-0 w-full">
                <div className="absolute bottom-0 w-full h-32 bg-green-700 rounded-t-full scale-x-150"></div>
                <div className="absolute bottom-0 left-1/4 w-full h-48 bg-green-600 rounded-t-full scale-x-125"></div>
            </div>

            {/* Main castles */}
            <div className="absolute bottom-0 w-full">
                {/* Blue castle (left) */}
                <div className="absolute bottom-28 left-8 w-48 h-64">
                    {/* Castle base */}
                    <div className="absolute bottom-0 w-full h-36 bg-blue-500 rounded-t-lg"></div>

                    {/* Castle towers */}
                    <div className="absolute bottom-36 left-2 w-10 h-20 bg-blue-500">
                        <div className="absolute top-0 left-0 w-10 h-4 bg-blue-500 flex justify-around">
                            <div className="w-2 h-4 bg-blue-500"></div>
                            <div className="w-2 h-4 bg-blue-500"></div>
                        </div>
                    </div>

                    <div className="absolute bottom-36 left-19 w-10 h-16 bg-blue-500">
                        <div className="absolute top-0 left-0 w-10 h-4 bg-blue-500 flex justify-around">
                            <div className="w-2 h-4 bg-blue-500"></div>
                            <div className="w-2 h-4 bg-blue-500"></div>
                        </div>
                    </div>

                    <div className="absolute bottom-36 right-2 w-10 h-20 bg-blue-500">
                        <div className="absolute top-0 left-0 w-10 h-4 bg-blue-500 flex justify-around">
                            <div className="w-2 h-4 bg-blue-500"></div>
                            <div className="w-2 h-4 bg-blue-500"></div>
                        </div>
                    </div>

                    <div className="absolute bottom-36 left-1/2 transform -translate-x-1/2 w-16 h-24 bg-blue-500">
                        <div className="absolute top-0 left-0 w-16 h-6 bg-blue-500 flex justify-around">
                            <div className="w-3 h-6 bg-blue-500"></div>
                            <div className="w-3 h-6 bg-blue-500"></div>
                            <div className="w-3 h-6 bg-blue-500"></div>
                        </div>
                    </div>

                    {/* Castle door */}
                    <div className="absolute bottom-0 left-1/2 transform -translate-x-1/2 w-12 h-16 bg-blue-800 rounded-t-lg"></div>

                    {/* Castle windows */}
                    <div className="absolute bottom-24 left-8 w-6 h-6 bg-blue-900 rounded-full"></div>
                    <div className="absolute bottom-24 right-8 w-6 h-6 bg-blue-900 rounded-full"></div>
                    <div className="absolute bottom-12 left-16 w-6 h-6 bg-blue-900 rounded-full"></div>
                    <div className="absolute bottom-12 right-16 w-6 h-6 bg-blue-900 rounded-full"></div>

                    {/* Flag */}
                    <div className="absolute bottom-60 left-1/2 transform -translate-x-1/2">
                        <div className="w-1 h-16 bg-gray-700"></div>
                        <div className={`w-8 h-6 bg-blue-400 absolute top-0 left-1`}
                            style={{ clipPath: `polygon(0 0, 100% ${flagWave <= 10 ? 20 : 80}%, 100% 100%, 0 ${flagWave <= 10 ? 80 : 20}%)` }}>
                        </div>
                    </div>
                </div>

                {/* Red castle (right) */}
                <div className="absolute bottom-28 right-8 w-48 h-64">
                    {/* Castle base */}
                    <div className="absolute bottom-0 w-full h-36 bg-red-500 rounded-t-lg"></div>

                    {/* Castle towers */}
                    <div className="absolute bottom-36 left-2 w-10 h-20 bg-red-500">
                        <div className="absolute top-0 left-0 w-10 h-4 bg-red-500 flex justify-around">
                            <div className="w-2 h-4 bg-red-500"></div>
                            <div className="w-2 h-4 bg-red-500"></div>
                        </div>
                    </div>

                    <div className="absolute bottom-36 left-19 w-10 h-16 bg-red-500">
                        <div className="absolute top-0 left-0 w-10 h-4 bg-red-500 flex justify-around">
                            <div className="w-2 h-4 bg-red-500"></div>
                            <div className="w-2 h-4 bg-red-500"></div>
                        </div>
                    </div>

                    <div className="absolute bottom-36 right-2 w-10 h-20 bg-red-500">
                        <div className="absolute top-0 left-0 w-10 h-4 bg-red-500 flex justify-around">
                            <div className="w-2 h-4 bg-red-500"></div>
                            <div className="w-2 h-4 bg-red-500"></div>
                        </div>
                    </div>

                    <div className="absolute bottom-36 left-1/2 transform -translate-x-1/2 w-16 h-24 bg-red-500">
                        <div className="absolute top-0 left-0 w-16 h-6 bg-red-500 flex justify-around">
                            <div className="w-3 h-6 bg-red-500"></div>
                            <div className="w-3 h-6 bg-red-500"></div>
                            <div className="w-3 h-6 bg-red-500"></div>
                        </div>
                    </div>

                    {/* Castle door */}
                    <div className="absolute bottom-0 left-1/2 transform -translate-x-1/2 w-12 h-16 bg-red-800 rounded-t-lg"></div>

                    {/* Castle windows */}
                    <div className="absolute bottom-24 left-8 w-6 h-6 bg-red-900 rounded-full"></div>
                    <div className="absolute bottom-24 right-8 w-6 h-6 bg-red-900 rounded-full"></div>
                    <div className="absolute bottom-12 left-16 w-6 h-6 bg-red-900 rounded-full"></div>
                    <div className="absolute bottom-12 right-16 w-6 h-6 bg-red-900 rounded-full"></div>

                    {/* Flag */}
                    <div className="absolute bottom-60 left-1/2 transform -translate-x-1/2">
                        <div className="w-1 h-16 bg-gray-700"></div>
                        <div className={`w-8 h-6 bg-red-400 absolute top-0 left-1`}
                            style={{ clipPath: `polygon(0 0, 100% ${flagWave <= 10 ? 20 : 80}%, 100% 100%, 0 ${flagWave <= 10 ? 80 : 20}%)` }}>
                        </div>
                    </div>
                </div>

                {/* Bridge connecting castles */}
                <div className="absolute bottom-12 left-1/2 transform -translate-x-1/2 w-64 h-8 bg-gray-700"></div>
                <div className="absolute bottom-12 left-1/2 transform -translate-x-1/2 w-64 h-2 bg-gray-900 flex justify-between px-4">
                    <div className="w-1 h-8 bg-gray-900"></div>
                    <div className="w-1 h-8 bg-gray-900"></div>
                    <div className="w-1 h-8 bg-gray-900"></div>
                    <div className="w-1 h-8 bg-gray-900"></div>
                    <div className="w-1 h-8 bg-gray-900"></div>
                </div>

                {/* River */}
                <div className="absolute bottom-0 left-1/2 transform -translate-x-1/2 w-72 h-12 bg-blue-300"></div>
            </div>

            {/* Decorative elements */}
            <div className="absolute bottom-32 left-1/2 transform -translate-x-1/2">
                {/* Game elements */}
                <div className="absolute -left-32 -top-8 w-8 h-8 bg-purple-500 rounded-md opacity-80"></div>
                <div className="absolute -right-36 -top-12 w-10 h-10 bg-cyan-500 rounded-md opacity-80"></div>
                <div className="absolute -left-48 -top-20 w-6 h-12 bg-green-500 opacity-80"></div>
            </div>

            {/* ROYAKA title */}
            <div className="z-10 text-center mb-8 mt-6">
                <h1 className="text-6xl font-extrabold text-cyan-400 drop-shadow-lg mb-2 select-none">
                    ROYAKA
                </h1>
                <p className="text-white text-xl font-semibold -mt-2 mb-12">Battle Arena</p>
            </div>

            {/* Single Play Game button */}
            <div className="z-10 mt-8">
                <button
                    onClick={handlePlay}
                    className="bg-cyan-500 text-white font-bold px-16 py-4 rounded-xl shadow-lg hover:bg-cyan-400 transition transform hover:scale-105 text-xl"
                >
                    Play Game
                </button>
            </div>

            {/* Game version */}
            <div className="absolute bottom-2 right-4 text-white text-opacity-50 text-sm">
                v1.0.0
            </div>
        </div>
    );
};

export default Entry;