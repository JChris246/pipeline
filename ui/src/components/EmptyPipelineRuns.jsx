const EmptyPipelineRuns = () => {
    return (
        <div className="flex flex-col items-center justify-center h-full text-center">
            <div className="w-24 h-24 bg-slate-700 rounded-full flex items-center justify-center mb-6">
                <svg className="w-12 h-12 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0
                        002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
            </div>
            <h3 className="text-2xl font-semibold text-slate-300 mb-2">No Pipeline Runs</h3>
            <p className="text-slate-400">This pipeline hasn't been executed yet</p>
        </div>
    );
};

export default EmptyPipelineRuns;