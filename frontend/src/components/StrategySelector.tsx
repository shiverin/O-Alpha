"use client";

import React, { useState, useEffect } from "react";

export interface StrategyConfig {
  strategy: string;
  qNoise: number;
  rNoise: number;
  zThresh: number;
  fastPeriod: number;
  slowPeriod: number;
}

interface StrategySelectorProps {
  onConfigChange: (config: StrategyConfig) => void;
}

const Slider = ({
  label,
  value,
  min,
  max,
  step,
  onChange,
  onCommit,
  prefix = "",
}: {
  label: string;
  value: number;
  min: string;
  max: string;
  step: string;
  onChange: (v: number) => void;
  onCommit: () => void;
  prefix?: string;
}) => {
  const minVal = parseFloat(min);
  const maxVal = parseFloat(max);
  const percent = ((value - minVal) / (maxVal - minVal)) * 100;

  return (
    <div className="group flex flex-col gap-3">
      <div className="flex justify-between items-end">
        <span className="text-sm font-light tracking-wide text-white/50 group-hover:text-white/70 transition-colors">
          {label}
        </span>
        <span className="text-sm font-mono text-white/90">
          {prefix}
          {value}
        </span>
      </div>

      <div className="relative flex items-center h-5">
        <input
          type="range"
          min={min}
          max={max}
          step={step}
          value={value}
          onChange={(e) => onChange(parseFloat(e.target.value))}
          onPointerUp={onCommit}
          className="absolute w-full h-[3px] appearance-none outline-none rounded-full cursor-pointer
            [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:h-4 
            [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-white 
            [&::-webkit-slider-thumb]:shadow-[0_0_10px_rgba(255,255,255,0.3)] 
            [&::-webkit-slider-thumb]:transition-[transform,box-shadow] [&::-webkit-slider-thumb]:duration-200 
            [&::-webkit-slider-thumb]:hover:scale-125 [&::-webkit-slider-thumb]:hover:shadow-[0_0_15px_rgba(255,255,255,0.6)]
            [&::-webkit-slider-thumb]:active:scale-90
            
            [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:w-4 [&::-moz-range-thumb]:h-4 
            [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:bg-white [&::-moz-range-thumb]:border-none
            [&::-moz-range-thumb]:shadow-[0_0_10px_rgba(255,255,255,0.3)] 
            [&::-moz-range-thumb]:transition-[transform,box-shadow] [&::-moz-range-thumb]:duration-200 
            [&::-moz-range-thumb]:hover:scale-125 [&::-moz-range-thumb]:hover:shadow-[0_0_15px_rgba(255,255,255,0.6)]
            [&::-moz-range-thumb]:active:scale-90"
          style={{
            background: `linear-gradient(to right, rgba(255,255,255,0.9) ${percent}%, rgba(255,255,255,0.1) ${percent}%)`,
          }}
        />
      </div>
    </div>
  );
};

export default function StrategySelector({
  onConfigChange,
}: StrategySelectorProps) {
  const [strategy, setStrategy] = useState("KALMAN");

  const [qNoise, setQNoise] = useState(0.01);
  const [rNoise, setRNoise] = useState(0.5);
  const [zThresh, setZThresh] = useState(2.0);
  const [fastPeriod, setFastPeriod] = useState(10);
  const [slowPeriod, setSlowPeriod] = useState(30);

  useEffect(() => {
    onConfigChange({
      strategy,
      qNoise,
      rNoise,
      zThresh,
      fastPeriod,
      slowPeriod,
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleUpdate = (updatedFields: Partial<StrategyConfig>) => {
    const nextState = {
      strategy,
      qNoise,
      rNoise,
      zThresh,
      fastPeriod,
      slowPeriod,
      ...updatedFields,
    };
    onConfigChange(nextState);
  };

  return (
    <div className="w-full bg-transparent flex flex-col gap-8 py-2">
      <div className="flex flex-col gap-4">
        <span className="text-[10px] uppercase tracking-[0.2em] text-white/40 font-medium">
          Simulation Engine
        </span>

        <div className="flex p-1 bg-white/5 rounded-lg border border-white/[0.05] backdrop-blur-xl">
          <button
            onClick={() => {
              setStrategy("KALMAN");
              handleUpdate({ strategy: "KALMAN" });
            }}
            className={`flex-1 text-xs py-2.5 rounded-md transition-all duration-300 ease-out font-medium tracking-wide ${
              strategy === "KALMAN"
                ? "bg-white/10 text-white shadow-sm"
                : "text-white/30 hover:text-white/60"
            }`}
          >
            Kalman Filter
          </button>
          <button
            onClick={() => {
              setStrategy("MA_CROSSOVER");
              handleUpdate({ strategy: "MA_CROSSOVER" });
            }}
            className={`flex-1 text-xs py-2.5 rounded-md transition-all duration-300 ease-out font-medium tracking-wide ${
              strategy === "MA_CROSSOVER"
                ? "bg-white/10 text-white shadow-sm"
                : "text-white/30 hover:text-white/60"
            }`}
          >
            Momentum
          </button>
        </div>
      </div>

      <div className="flex flex-col gap-8 min-h-[220px]">
        {strategy === "KALMAN" && (
          <div className="flex flex-col gap-8 animate-in fade-in zoom-in-95 duration-500 ease-out">
            <Slider
              label="Process Noise"
              min="0.001"
              max="0.1"
              step="0.001"
              value={qNoise}
              onChange={(v) => setQNoise(v)}
              onCommit={() => handleUpdate({ qNoise })}
            />
            <Slider
              label="Noise Rejection"
              min="0.1"
              max="5.0"
              step="0.1"
              value={rNoise}
              onChange={(v) => setRNoise(v)}
              onCommit={() => handleUpdate({ rNoise })}
            />
            <Slider
              label="Execution Threshold"
              min="1.0"
              max="4.0"
              step="0.1"
              value={zThresh}
              prefix="±"
              onChange={(v) => setZThresh(v)}
              onCommit={() => handleUpdate({ zThresh })}
            />
          </div>
        )}

        {strategy === "MA_CROSSOVER" && (
          <div className="flex flex-col gap-8 animate-in fade-in zoom-in-95 duration-500 ease-out">
            <Slider
              label="Fast MA Window"
              min="2"
              max="50"
              step="1"
              value={fastPeriod}
              onChange={(v) => setFastPeriod(v)}
              onCommit={() => handleUpdate({ fastPeriod })}
            />
            <Slider
              label="Slow MA Window"
              min="10"
              max="200"
              step="1"
              value={slowPeriod}
              onChange={(v) => setSlowPeriod(v)}
              onCommit={() => handleUpdate({ slowPeriod })}
            />
          </div>
        )}
      </div>
    </div>
  );
}
