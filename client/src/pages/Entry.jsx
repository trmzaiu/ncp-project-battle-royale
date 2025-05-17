import { useState, useEffect } from 'react';
import { useNavigate } from "react-router-dom";

export default function RoyakaBattleArena() {
  const [animationComplete, setAnimationComplete] = useState(false);
  const [showTitle, setShowTitle] = useState(false);
  const [showButton, setShowButton] = useState(false);
  const navigate = useNavigate();
  const handlePlay = () => {
    console.log("Navigating to auth page...");
    setTimeout(() => navigate("/auth"), 1500);
  };
  
  useEffect(() => {
    // Sequence the animations with more time for slower connections
    const timer1 = setTimeout(() => setAnimationComplete(true), 600);
    const timer2 = setTimeout(() => setShowTitle(true), 1400);
    const timer3 = setTimeout(() => setShowButton(true), 2000);
    
    return () => {
      clearTimeout(timer1);
      clearTimeout(timer2);
      clearTimeout(timer3);
    };
  }, []);

  return (
    <div className="flex flex-col items-center justify-center min-h-screen w-full bg-gradient-to-b from-blue-900 via-blue-800 to-blue-900 overflow-hidden relative">
      {/* Animated background elements - with better positioning */}
      <div className="absolute inset-0 overflow-hidden">
        {/* Background particles */}
        <div className="absolute w-full h-full">
          {[...Array(20)].map((_, i) => (
            <div 
              key={i}
              className="absolute w-1 h-1 bg-blue-300 rounded-full animate-pulse"
              style={{
                top: `${Math.random() * 100}%`,
                left: `${Math.random() * 100}%`,
                opacity: 0.4,
                animationDuration: `${Math.random() * 3 + 2}s`,
                animationDelay: `${Math.random() * 2}s`
              }}
            ></div>
          ))}
        </div>
      </div>
      
      {/* Main content container - with responsive width */}
      <div className="z-10 relative w-full min-h-screen max-w-xs sm:max-w-sm md:max-w-md lg:max-w-lg xl:max-w-2xl px-4 sm:px-6 py-8 sm:py-12">
        {/* Crown animation that comes from top - responsive sizing */}
        <div className="relative flex justify-center mb-2 sm:mb-4">
          <div 
            className={`transition-all duration-1000 ease-out transform
                      ${animationComplete ? 'translate-y-0 opacity-100' : '-translate-y-24 opacity-0'}`}
          >
            <svg 
              className="w-24 h-24 sm:w-32 sm:h-32 md:w-40 md:h-40" 
              viewBox="0 0 512 512" 
              fill="none" 
              xmlns="http://www.w3.org/2000/svg"
            >
              {/* Crown base */}
              <path d="M76 352L128 208L216 272L256 176L296 272L384 208L436 352H76Z" fill="#FFC107" />
              <path d="M96 352L144 224L216 280L256 192L296 280L368 224L416 352H96Z" fill="#FFD54F" />
              
              {/* Crown bottom band */}
              <path d="M76 352H436V392H76V352Z" fill="#FFA000" />
              <path d="M96 352H416V382H96V352Z" fill="#FFB300" />
              
              {/* Crown spikes */}
              <path d="M116 352V312L144 336L172 312V352H116Z" fill="#FFD54F" />
              <path d="M172 352V312L200 336L228 312V352H172Z" fill="#FFD54F" />
              <path d="M228 352V312L256 336L284 312V352H228Z" fill="#FFD54F" />
              <path d="M284 352V312L312 336L340 312V352H284Z" fill="#FFD54F" />
              <path d="M340 352V312L368 336L396 312V352H340Z" fill="#FFD54F" />
              
              {/* Crown jewels */}
              <circle cx="144" cy="244" r="12" fill="#F44336" />
              <circle cx="256" cy="208" r="16" fill="#2196F3" />
              <circle cx="368" cy="244" r="12" fill="#4CAF50" />
              
              {/* Crown highlights */}
              <path d="M200 280L216 288L256 200L296 288L312 280L256 180L200 280Z" fill="#FFECB3" />
            </svg>
          </div>
          
          {/* Flash effect when crown lands - responsive sizing */}
          <div 
            className={`absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2
                      transition-all duration-500 delay-800
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
        
        {/* ROYAKA logo text with animated reveal - responsive text sizing */}
        <div className={`relative transition-all duration-700 ${showTitle ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-10'}`}>
          <div className="text-center">
            {/* Main title with blue-to-yellow gradient similar to Clash Royale */}
            <h1 className="text-4xl sm:text-5xl md:text-6xl font-extrabold text-transparent bg-clip-text bg-gradient-to-b from-blue-200 via-blue-100 to-yellow-200 mb-2 sm:mb-3 tracking-wide select-none drop-shadow-lg">
              ROYAKA
            </h1>
            
            {/* Battle Arena subtitle with animated reveal */}
            <div className="relative overflow-hidden">
              <p className="text-white text-lg sm:text-xl font-semibold mb-2 sm:mb-3">Battle Arena</p>
              
              {/* Gold accent bar - responsive width */}
              <div className="h-1 w-24 sm:w-32 bg-gradient-to-r from-yellow-400 to-yellow-600 mx-auto rounded-full"></div>
            </div>
          </div>
        </div>
        
        {/* Blue light beam effects from top and bottom - responsive sizing */}
        <div className="absolute inset-0 pointer-events-none overflow-hidden">
          <div className={`absolute left-1/2 top-0 w-24 sm:w-32 h-64 sm:h-96 bg-blue-500 opacity-20 rounded-full blur-xl transform -translate-x-1/2 -translate-y-1/2 transition-opacity duration-1000 ${showTitle ? 'opacity-20' : 'opacity-0'}`}></div>
          <div className={`absolute left-1/2 bottom-0 w-24 sm:w-32 h-64 sm:h-96 bg-blue-500 opacity-20 rounded-full blur-xl transform -translate-x-1/2 translate-y-1/2 transition-opacity duration-1000 ${showTitle ? 'opacity-20' : 'opacity-0'}`}></div>
        </div>
        
        {/* Play Game button with Clash Royale style - responsive sizing */}
        <div className={`z-10 mt-6 sm:mt-8 md:mt-10 text-center transition-all duration-700 ${showButton ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-10'}`}>
          <button
            onClick={handlePlay}
            className="relative overflow-hidden group bg-gradient-to-b from-yellow-400 to-yellow-600 text-white font-bold px-8 sm:px-12 md:px-16 py-3 sm:py-4 rounded-xl shadow-lg border-b-4 border-yellow-700 transform transition-all duration-200 hover:scale-105 active:translate-y-1 active:border-b-2"
          >
            <span className="relative z-10 text-base sm:text-lg md:text-xl uppercase tracking-wider">Play Now</span>
            
            {/* Button shine effect */}
            <span className="absolute top-0 -left-full h-full w-1/2 bg-white opacity-20 transform skew-x-12 group-hover:translate-x-[250%] transition-all duration-700"></span>
          </button>
        </div>
        
        {/* Bottom decorative elements - similar to cards - hide on smallest screens */}
        <div className={`absolute bottom-20 left-0 w-full flex justify-between px-4 sm:px-8 pb-2 sm:pb-4 transition-all duration-1000 ${showButton ? 'opacity-100' : 'opacity-0'} hidden sm:flex`}>
          <div className="w-12 h-16 sm:w-16 sm:h-20 bg-gradient-to-br from-blue-700 to-blue-900 rounded-lg shadow-lg transform -rotate-12 opacity-40"></div>
          <div className="w-12 h-16 sm:w-16 sm:h-20 bg-gradient-to-br from-red-700 to-red-900 rounded-lg shadow-lg transform rotate-12 opacity-40"></div>
        </div>
      </div>
      
      {/* Mobile portrait mode optimization */}
      <div className="absolute bottom-6 left-0 w-full text-center text-xs text-blue-200 opacity-60 px-4 md:hidden">
        <p>Best experienced in landscape mode</p>
      </div>
    </div>
  );
}