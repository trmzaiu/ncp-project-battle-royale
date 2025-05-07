# Clash of Realms: Text Royale

A networked, text-based multiplayer battle game inspired by Clash Royale. This game emphasizes strategy, leveling, and real-time decision-making over TCP/UDP communication protocols.

---

## 🎯 Game Objectives

- Destroy the opponent’s **King Tower** before the 3-minute timer ends.
- If the King Tower survives, the player who destroys more **Guard Towers** wins.
- Gain **EXP**, **level up troops/towers**, and climb the ranks!

---

## 🏰 Basic Game Rules

### 1. **Player Connection**
- Two players connect via TCP/UDP with **username/password**.

### 2. **Towers & Troops**
- Each player has:
  - **1 King Tower**
  - **2 Guard Towers**
  - **3 Troops**, spawned randomly
- Each troop/tower has:
  - **HP**, **ATK**, **DEF**
  - Optional: **CRIT chance**

### 3. **Turn-Based (Simple Version)**
- Players take turns deploying troops.
- Destroyed troops or towers are removed from the game.

### 4. **Damage Formula**
```
DMG = ATK_A - DEF_B (if ≥ 0)
HP_B = HP_B - DMG
```

---

## ⚡ Enhanced Mechanics

### 1. **Critical Hits**
- 10% CRIT chance = 1.2x ATK
```
DMG = (ATK or ATK * 1.2 if CRIT) - DEF
```

### 2. **Continuous Play (No Turns)**
- Game lasts **3 minutes**, players attack simultaneously.

### 3. **Mana System**
- Starts at 5 MANA, +1/sec regen, max 10.
- Troops cost MANA to summon.

### 4. **EXP System**
- Win: +30 EXP, Draw: +10 EXP
- EXP = troop/tower stat +10% per level.
- EXP required per level increases by 10%.

---

## 🧠 New Strategic Systems

### 5. **Troop Type Advantage**
- Infantry > Archer > Cavalry > Infantry
- +20% ATK if advantaged, -10% if disadvantaged

### 6. **Weather Effects**
Randomly selected at start of game:
- ☀️ Sun: +10% Mana Regen
- 🌧️ Rain: -20% DEF
- 🌪️ Wind: +10% CRIT Chance

### 7. **Troop Skills**
- **Knight**: 20% chance to block 50% DMG
- **Prince**: +50% damage to Guard Towers
- **Queen**: Heals the lowest HP tower by 300

### 8. **Tower Upgrades**
Spend EXP between matches:
- +10% HP → 200 EXP
- +10% ATK → 300 EXP
- +5% CRIT → 500 EXP

### 9. **Random Events**
Occurs every 30s:
- 🔥 Firestorm: Troops lose 10% HP
- ⚡ Overload: +50% ATK for 5s
- 🛡️ Reinforce: Guard Towers +200 HP

### 10. **Power-Ups**
Players can equip 1–2 before match:
- Mana Potion: +3 Mana instantly
- Revive Scroll: Revive one troop
- Shield Orb: Block the next attack

### 11. **Leaderboard & Ranks**
- Ranks: Bronze, Silver, Gold, Diamond
- EXP-based progression and scoreboard



