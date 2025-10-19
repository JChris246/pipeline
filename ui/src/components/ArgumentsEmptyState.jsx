const ArgumentsEmptyState = () => {
    return (
        <div className="text-center py-4 text-slate-400 text-sm bg-slate-800/30 rounded-lg border border-slate-700/30">
            <div className="w-6 h-6 mx-auto mb-2 text-slate-500">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                        d="M8 9l3 3-3 3m5 0h3" />
                </svg>
            </div>
            <p>No arguments added</p>
            <p className="text-xs text-slate-500 mt-1">Click "Add Arg" to add command line arguments</p>
        </div>
    );
};

export default ArgumentsEmptyState;
