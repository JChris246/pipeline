import { SettingsIcon } from "../icons";

const EnvVarsEmptyState = () => {
    return (
        <div className="text-center py-4 text-slate-400 text-sm bg-slate-800/30 rounded-lg border border-slate-700/30">
            <div className="w-6 h-6 mx-auto mb-2 text-slate-500">
                <SettingsIcon className="w-6 h-6" />
            </div>
            <p>No environment variables added</p>
            <p className="text-xs text-slate-500 mt-1">Click "Add Env Var" to add environment variables for this stage</p>
        </div>
    );
};

export default EnvVarsEmptyState;
