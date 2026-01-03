import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { Activity, Shield, Zap, CheckCircle, Flame, Play, RefreshCw, Server } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

interface ProviderStatus {
    name: string;
    healthy: boolean;
    latency_ms: number;
    breaker_state: string;
    cost_per_req: number;
}

interface SystemStatus {
    providers: ProviderStatus[];
    timestamp: number;
}

const API_BASE_URL = 'http://localhost:8080';

const App: React.FC = () => {
    const [status, setStatus] = useState<SystemStatus | null>(null);
    const [history, setHistory] = useState<any[]>([]);
    const [error, setError] = useState<string | null>(null);
    const [lastTestResult, setLastTestResult] = useState<any>(null);
    const [loading, setLoading] = useState<boolean>(false);

    const fetchData = async () => {
        try {
            const response = await axios.get(`${API_BASE_URL}/api/v1/status`);
            setStatus(response.data);
            setError(null);
            if (response.data.providers && response.data.providers.length > 0) {
                setHistory(prev => [...prev.slice(-19), {
                    time: new Date().toLocaleTimeString(),
                    latency: response.data.providers.reduce((acc: any, p: any) => acc + p.latency_ms, 0) / response.data.providers.length
                }]);
            }
        } catch (err) {
            console.error("Failed to fetch status:", err);
            setError("Disconnected from Backend");
        }
    };

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 2000);
        return () => clearInterval(interval);
    }, []);

    const runTestRequest = async () => {
        setLoading(true);
        try {
            const response = await axios.post(`${API_BASE_URL}/api/v1/test-rpc`);
            setLastTestResult(response.data);
            fetchData();
        } catch (err) {
            setLastTestResult({ error: "Request Failed" });
        }
        setLoading(false);
    };

    const tripProvider = async (name: string) => {
        try {
            await axios.post(`${API_BASE_URL}/api/v1/chaos/trip?provider=${name}`);
            fetchData();
        } catch (err) {
            console.error("Failed to trip provider", err);
        }
    };

    const resetChaos = async () => {
        try {
            await axios.post(`${API_BASE_URL}/api/v1/chaos/reset`);
            fetchData();
        } catch (err) {
            console.error("Failed to reset chaos", err);
        }
    };

    return (
        <div className="min-h-screen p-8 max-w-7xl mx-auto space-y-8">
            {/* Header */}
            <header className="flex justify-between items-center mb-12">
                <div>
                    <h1 className="text-4xl font-extrabold tracking-tight flex items-center gap-3">
                        <Shield className="w-10 h-10 text-primary" />
                        Heimdall <span className="text-primary">S3R</span>
                    </h1>
                    <p className="text-white/50 mt-2 font-medium">Smart RPC Reliability Router Dashboard</p>
                </div>
                <div className="flex items-center gap-4">
                    <button onClick={resetChaos} className="glass px-4 py-2 hover:bg-white/10 transition-all flex items-center gap-2 text-sm">
                        <RefreshCw className="w-4 h-4" />
                        Reset All Systems
                    </button>
                    <div className="glass px-4 py-2 flex items-center gap-2 border-primary/30">
                        <Activity className={`w-4 h-4 ${error ? 'text-danger animate-pulse' : 'text-success'}`} />
                        <span className="text-sm font-bold uppercase tracking-wider">{error || 'Live'}</span>
                    </div>
                </div>
            </header>

            {/* Top Stats Row */}
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
                <div className="lg:col-span-3 glass p-6">
                    <div className="flex justify-between items-center mb-6">
                        <h2 className="text-xl font-bold flex items-center gap-2">
                            <Zap className="w-5 h-5 text-warning" />
                            Global Latency Benchmarks
                        </h2>
                    </div>
                    <div className="h-[300px] w-full mt-4">
                        <ResponsiveContainer width="100%" height="100%">
                            <LineChart data={history}>
                                <CartesianGrid strokeDasharray="3 3" stroke="#ffffff10" vertical={false} />
                                <XAxis dataKey="time" stroke="#ffffff30" fontSize={11} tickLine={false} axisLine={false} />
                                <YAxis stroke="#ffffff30" fontSize={11} tickLine={false} axisLine={false} tickFormatter={(v) => `${v}ms`} />
                                <Tooltip
                                    contentStyle={{ backgroundColor: '#141417', border: '1px solid #ffffff10', borderRadius: '8px', fontSize: '12px' }}
                                    itemStyle={{ color: '#3b82f6' }}
                                />
                                <Line type="monotone" dataKey="latency" stroke="#3b82f6" strokeWidth={3} dot={false} animationDuration={300} />
                            </LineChart>
                        </ResponsiveContainer>
                    </div>
                </div>

                {/* Command Panel */}
                <div className="glass p-6 border-t-4 border-warning">
                    <h2 className="text-xl font-bold flex items-center gap-2 mb-6 text-warning">
                        <Play className="w-5 h-5" />
                        Request Lab
                    </h2>
                    <div className="space-y-6">
                        <button
                            disabled={loading}
                            onClick={runTestRequest}
                            className={`w-full py-4 rounded-xl font-bold transition-all flex flex-col items-center justify-center gap-2 shadow-lg shadow-primary/20 bg-primary hover:bg-primary/80 active:scale-95 text-white ${loading ? 'opacity-50' : ''}`}
                        >
                            {loading ? <RefreshCw className="animate-spin" /> : <Zap className="w-6 h-6" />}
                            Fire RPC Request
                        </button>

                        {lastTestResult && (
                            <div className="bg-black/40 rounded-lg p-4 font-mono text-xs overflow-hidden border border-white/5">
                                <p className="text-primary mb-1 font-bold tracking-widest uppercase text-[10px]">Last Result</p>
                                <div className="space-y-1">
                                    <p><span className="text-white/40">Provider:</span> <span className="text-success capitalize font-bold">{lastTestResult.provider}</span></p>
                                    <p className="truncate text-white/60"><span className="text-white/40">Response ID:</span> {lastTestResult.response?.id || 'N/A'}</p>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            {/* Providers Row */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {status?.providers ? (
                    status.providers.map(p => (
                        <div key={p.name} className="glass p-6 flex flex-col justify-between group">
                            <div className="flex justify-between items-start mb-6">
                                <div>
                                    <h3 className="text-2xl font-black capitalize tracking-tight">{p.name}</h3>
                                    <div className={`mt-1 flex items-center gap-1.5 px-2 py-0.5 rounded text-[10px] font-bold uppercase ${p.breaker_state.includes('OPEN') ? 'bg-danger/20 text-danger' : 'bg-success/20 text-success'}`}>
                                        <Server className="w-3 h-3" />
                                        {p.breaker_state}
                                    </div>
                                </div>
                                <div className="text-right">
                                    <p className={`text-2xl font-black font-mono ${p.latency_ms > 150 ? 'text-danger' : 'text-success'}`}>
                                        {p.latency_ms}<span className="text-xs ml-0.5">ms</span>
                                    </p>
                                    <p className="text-[10px] font-bold text-white/30 tracking-widest uppercase">Latency</p>
                                </div>
                            </div>

                            <div className="space-y-4">
                                <div className="flex justify-between text-xs font-bold uppercase tracking-widest text-white/40">
                                    <span>Reliability</span>
                                    <span>{p.healthy ? '100%' : '0%'}</span>
                                </div>
                                <div className="w-full h-1.5 bg-white/5 rounded-full overflow-hidden">
                                    <div className={`h-full transition-all duration-500 ${p.healthy ? 'bg-success w-full pulse-success' : 'bg-danger w-0'}`} />
                                </div>

                                <button
                                    onClick={() => tripProvider(p.name)}
                                    className="w-full mt-4 py-3 glass hover:bg-danger hover:text-white transition-all text-xs font-bold uppercase tracking-widest text-danger flex items-center justify-center gap-2 group-hover:scale-[1.02]"
                                >
                                    <Flame className="w-4 h-4" />
                                    Simulate Failure
                                </button>
                            </div>
                        </div>
                    ))
                ) : (
                    <div className="lg:col-span-3 text-center py-20 glass text-white/30 italic">
                        Waiting for orchestrator...
                    </div>
                )}
            </div>

            {/* Footer Features */}
            <section className="grid grid-cols-1 md:grid-cols-3 gap-6 opacity-80 hover:opacity-100 transition-opacity">
                <div className="p-1 rounded-xl bg-gradient-to-r from-primary/50 to-transparent">
                    <div className="glass h-full p-6">
                        <h3 className="font-bold text-lg mb-2 flex items-center gap-2"><Zap className="text-primary w-5 h-5" /> Least-Latency</h3>
                        <p className="text-sm text-white/50 leading-relaxed font-medium">Heimdall continuously benchmarks and routes traffic to the fastest RPC node available.</p>
                    </div>
                </div>
                <div className="p-1 rounded-xl bg-gradient-to-r from-success/50 to-transparent">
                    <div className="glass h-full p-6">
                        <h3 className="font-bold text-lg mb-2 flex items-center gap-2"><CheckCircle className="text-success w-5 h-5" /> Redis Caching</h3>
                        <p className="text-sm text-white/50 leading-relaxed font-medium">Sub-millisecond response times for cached slots, slashing compute costs by up to 80%.</p>
                    </div>
                </div>
                <div className="p-1 rounded-xl bg-gradient-to-r from-warning/50 to-transparent">
                    <div className="glass h-full p-6">
                        <h3 className="font-bold text-lg mb-2 flex items-center gap-2"><Shield className="text-warning w-5 h-5" /> Self-Healing</h3>
                        <p className="text-sm text-white/50 leading-relaxed font-medium">Automatic circuit breaker isolation ensures failure cascading never reaches your users.</p>
                    </div>
                </div>
            </section>
        </div>
    );
};

export default App;
