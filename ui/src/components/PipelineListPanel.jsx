import { useState, useEffect, useRef } from "react";
import { request } from "../utils/Fetch";
import { AddIcon,  ClockIcon, CloseIcon, DeleteIcon, EditIcon, LightningIcon, PipelineIcon, PlayIcon,
    RefreshIcon, SearchIcon, SkipIcon, UploadIcon } from "../icons";

import { useAppContext } from "../AppContext";
import { NotificationType, useNotificationContext } from "./Notification";
import { getRelativeTime } from "../utils/utils";

import Modal from "./Modal";
import PipelineEmptyState from "./PipelineEmptyState";
import ArgumentsEmptyState from "./ArgumentsEmptyState";
import EnvVarsEmptyState from "./EnvVarsEmptyState";

const stageInitialState = {
    name: "",
    task: "",
    args: [],
    depends_on: [],
    pwd: "",
    skip: false,
    env: []
};

const pipelineInitialState = {
    name: "",
    parallel: false,
    stages: [JSON.parse(JSON.stringify(stageInitialState))],
    variables: [[]]
};

const HIGH_LEVEL_STATUS = 14;
const status = { IDLE: 0, COMPLETE: 1, FAILED: 2, RUNNING: 3 };

const formatLastRun = (timestamp) => {
    if (!timestamp) return "Never";

    return getRelativeTime(timestamp);
};

const PipelineListPanel = () => {
    const { display: displayNotification } = useNotificationContext();
    const { pipelines, setPipelines, setSelectedPipeline, setShowDetails } = useAppContext();
    const [searchFilter, setSearchFilter] = useState("");
    const [showPipelineModal, setShowPipelineModal] = useState(false);
    const [pipeline, setPipeline] = useState(pipelineInitialState);
    const [existing, setExisting] = useState(false);

    const pipelinesContainer = useRef();

    const getPipelines = () => {
        request({ url: "/api/pipelines",
            callback: ({ msg, success, json }) => {
                if (success) {
                    setPipelines(json);
                } else {
                    displayNotification({ message: "An error occurred fetching pipelines: " + msg, type: NotificationType.Error });
                }
            }
        });
    };

    // I do as I like
    // eslint-disable-next-line react-hooks/exhaustive-deps
    useEffect(() => getPipelines(), []);

    const showDetails = (name) => {
        setSelectedPipeline(name);
        setShowDetails(true);
    };

    const validateStages = () => {
        if (pipeline.stages.length < 1) {
            displayNotification({ message: "Pipeline must have at least one stage", type: NotificationType.Error });
            return false;
        }

        for (let i = 0; i < pipeline.stages.length; i++) {
            if (pipeline.stages[i].name.trim().length < 1) {
                displayNotification({ message: "Stage name cannot be empty", type: NotificationType.Error });
                return false;
            }

            if (pipeline.stages[i].task.trim().length < 1) {
                displayNotification({ message: "Stage task cannot be empty", type: NotificationType.Error });
                return false;
            }

            // TODO: validate dependencies
            // TODO: validate variables
        }

        return true;
    };

    const createPipelineRequestPayload = () => {
        const variables = {};
        for (let i = 0; i < pipeline.variables.length; i++) {
            if (pipeline.variables[i].length < 2 || pipeline.variables[i][0].trim().length < 1 || pipeline.variables[i][1].trim().length < 1) {
                continue;
            }
            variables[pipeline.variables[i][0].trim()] = pipeline.variables[i][1].trim();
        }

        return JSON.stringify({
            pipeline: {
                name: pipeline.name.trim(),
                parallel: pipeline.parallel,
                stages: pipeline.stages,
            },
            variables
        });
    };

    const editPipelineRequestPayload = () => {
        const variables = {};
        for (let i = 0; i < pipeline.variables.length; i++) {
            if (pipeline.variables[i].length < 2 || pipeline.variables[i][0].trim().length < 1 || pipeline.variables[i][1].trim().length < 1) {
                continue;
            }
            variables[pipeline.variables[i][0].trim()] = pipeline.variables[i][1].trim();
        }

        return JSON.stringify({
            name: pipeline.name.trim(),
            parallel: pipeline.parallel,
            stages: pipeline.stages,
            variables
        });
    };

    const addVariable = e => {
        e.preventDefault();
        setPipeline({ ...pipeline, variables: [...pipeline.variables, []] });
        e.target.blur();
    };

    const addStage = e => {
        e.preventDefault();
        setPipeline({ ...pipeline, stages: [...pipeline.stages, JSON.parse(JSON.stringify(stageInitialState))] });
        e.target.blur();
    };

    const deleteStage = (e, index) => {
        e.preventDefault();
        const newStages = [...pipeline.stages];
        newStages.splice(index, 1);
        setPipeline({ ...pipeline, stages: newStages });
        e.target.blur();
    };

    const updateVariable = (index, str, isKey) => {
        const newVariables = [...pipeline.variables];
        newVariables[index][isKey ? 0 : 1] = str;
        setPipeline({ ...pipeline, variables: newVariables });
    };

    const updateStage = (index, str, key) => {
        const newStages = [...pipeline.stages];
        newStages[index][key] = str;
        setPipeline({ ...pipeline, stages: newStages });
    };

    const addDependency = (e, index) => {
        const { value } = e.target;
        if (!value) {
            return;
        }

        const newStages = [...pipeline.stages];
        newStages[index].depends_on.push(value);
        setPipeline({ ...pipeline, stages: newStages });
        e.target.blur();
    };

    const removeDependency = (e, index, depIndex) => {
        const newStages = [...pipeline.stages];
        newStages[index].depends_on.splice(depIndex, 1);
        setPipeline({ ...pipeline, stages: newStages });
        e.target.blur();
    };

    const addArg = (e, index) => {
        e.preventDefault();
        const newStages = [...pipeline.stages];
        newStages[index].args.push("");
        setPipeline({ ...pipeline, stages: newStages });
        e.target.blur();
    };

    const updateArg = (stageIndex, argIndex, value) => {
        const newStages = [...pipeline.stages];
        newStages[stageIndex].args[argIndex] = value;
        setPipeline({ ...pipeline, stages: newStages });
    };

    const removeArg = (e, stageIndex, argIndex) => {
        e.preventDefault();
        const newStages = [...pipeline.stages];
        newStages[stageIndex].args.splice(argIndex, 1);
        setPipeline({ ...pipeline, stages: newStages });
        e.target.blur();
    };

    const addEnvVar = (e, index) => {
        e.preventDefault();
        const newStages = [...pipeline.stages];
        newStages[index].env.push("");
        setPipeline({ ...pipeline, stages: newStages });
        e.target.blur();
    };

    const updateEnvVar = (stageIndex, envIndex, value) => {
        const newStages = [...pipeline.stages];
        newStages[stageIndex].env[envIndex] = value;
        setPipeline({ ...pipeline, stages: newStages });
    };

    const removeEnvVar = (e, stageIndex, envIndex) => {
        e.preventDefault();
        const newStages = [...pipeline.stages];
        newStages[stageIndex].env.splice(envIndex, 1);
        setPipeline({ ...pipeline, stages: newStages });
        e.target.blur();
    };

    const showAddPipelineDialog = () => {
        setPipeline({ ...pipelineInitialState });
        setExisting(false);
        setShowPipelineModal(true);
    };

    const showEditPipelineDialog = (e, pipelineName) => {
        e.stopPropagation();

        request({ url: "/api/pipelines/" + pipelineName,
            callback: ({ msg, success, json }) => {
                if (success) {
                    setPipeline({
                        name: json.name,
                        parallel: json.parallel,
                        stages: json.stages,
                        variables: Object.entries(json.variables)
                    });
                    setExisting(true);
                    setShowPipelineModal(true);
                } else {
                    displayNotification({ message: "An error occurred fetching pipeline details: " + msg, type: NotificationType.Error });
                }
            }
        });
    };

    const addPipeline = e => {
        e.preventDefault();
        e.target.blur();

        if (pipeline.name.trim().length < 1) {
            displayNotification({ message: "Pipeline name cannot be empty", type: NotificationType.Error });
            return;
        }

        if (!validateStages()) {
            return;
        }

        request({ url: "/api/pipelines/register/json", method: "POST", body: createPipelineRequestPayload(),
            callback: ({ msg, success }) => {
                if (success) {
                    setShowPipelineModal(false);
                    displayNotification({ message: "Pipeline registered", type: NotificationType.Success });
                    getPipelines();
                } else {
                    displayNotification({ message: "An error occurred adding pipeline: " + msg, type: NotificationType.Error });
                }
            }
        });
    };

    const updatePipeline = e => {
        e.preventDefault();
        e.target.blur();

        // temporarily don't not allow editing pipeline name

        if (!validateStages()) {
            return;
        }

        request({ url: "/api/pipelines/" + pipeline.name.trim(), method: "PATCH", body: editPipelineRequestPayload(),
            callback: ({ msg, success }) => {
                if (success) {
                    setShowPipelineModal(false);
                    displayNotification({ message: "Pipeline updated", type: NotificationType.Success });
                    getPipelines();
                } else {
                    displayNotification({ message: "An error occurred updating pipeline: " + msg, type: NotificationType.Error });
                }
            }
        });
    };

    const deletePipeline = (e, pipelineName) => {
        e.preventDefault();
        e.target.blur();
        setShowPipelineModal(false);

        request({ url: "/api/pipelines/register/" + pipelineName,
            method: "DELETE",
            callback: ({ json, msg, success }) => {
                if (success) {
                    // TODO: if pipeline runs is open for this pipeline close it
                    setShowPipelineModal(false);
                    displayNotification({ message: json.msg, type: NotificationType.Info });
                    getPipelines();
                } else {
                    displayNotification({ message: "An error occurred deleting pipeline: " + msg, type: NotificationType.Error });
                }
            }
        });
    };

    const pipelineSorter = (a, b) => {
        if (a.status && b.status) {

            const aStatusLevel = !status[a.status.toUpperCase()] &&
                status[a.status.toUpperCase()] !== 0 ? HIGH_LEVEL_STATUS : status[a.status.toUpperCase()];
            const bStatusLevel = !status[b.status.toUpperCase()] &&
                status[b.status.toUpperCase()] !== 0 ? HIGH_LEVEL_STATUS : status[b.status.toUpperCase()];

            if (aStatusLevel === bStatusLevel) {
                return a.name?.localeCompare(b.name);
            }

            return bStatusLevel - aStatusLevel;
        }
        return a.name?.localeCompare(b.name);
    };

    const pipelineFilter = ({ name }) => searchFilter === "" || name.toLowerCase().includes(searchFilter.toLowerCase());

    const addPipelineTemplate = () => {
        return <div className="w-full h-full lg:w-3/5 lg:h-5/6 mx-auto shadow-sm bg-gradient-to-br from-slate-800/95 to-slate-900/95
            backdrop-blur-xl lg:rounded-3xl flex flex-col border border-slate-600/30 overflow-hidden">
            <div className="w-full p-8 flex justify-between items-center border-b border-slate-600/30
                bg-gradient-to-r from-slate-700/50 to-slate-800/50">
                <div className="flex items-center space-x-3">
                    <div className="w-12 h-12 rounded-2xl bg-gradient-to-r from-blue-600 to-indigo-600 flex items-center justify-center">
                        <PipelineIcon className="w-6 h-6 text-white" />
                    </div>
                    <div>
                        <h2 className="text-2xl font-bold text-slate-100">{existing ? "Edit Pipeline" : "Create Pipeline"}</h2>
                        <p className="text-sm text-slate-400">Configure your pipeline settings and stages</p>
                    </div>
                </div>
                <button onClick={() => setShowPipelineModal(false)}
                    className="w-12 h-12 rounded-2xl bg-slate-700/50 hover:bg-red-500/20 focus:bg-red-500/20 transition-all
                        duration-200 flex items-center justify-center text-slate-400 hover:text-red-400 border
                        border-slate-600/30 hover:border-red-500/30">
                    <CloseIcon className="w-5 h-5" />
                </button>
            </div>

            {/* Body */}
            <form className="flex flex-col justify-between p-8 h-full overflow-y-auto">
                <div className="space-y-8">
                    <div className="space-y-3">
                        <label htmlFor="pipelineName" className="block text-sm font-semibold text-slate-200">
                            Pipeline Name
                            <span className="text-red-400 ml-1">*</span>
                        </label>
                        <input type="text" placeholder="Enter a unique pipeline name" name="pipelineName" value={pipeline.name}
                            onChange={(e) => setPipeline({ ...pipeline, name: e.target.value })} disabled={existing}
                            className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-xl text-slate-100
                                placeholder-slate-400 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                        />
                    </div>

                    <div className="space-y-3">
                        <label htmlFor="pipelineParallel" className="block text-sm font-semibold text-slate-200">Execution Mode</label>
                        <div className="flex items-center justify-between p-4 bg-slate-700/30 rounded-xl border border-slate-600/30">
                            <div className="flex items-center space-x-3">
                                <div className="w-10 h-10 rounded-xl bg-gradient-to-r from-green-600 to-emerald-600 flex items-center justify-center">
                                    <LightningIcon className="w-5 h-5 text-white" />
                                </div>
                                <div>
                                    <p className="font-medium text-slate-200">Parallel Execution</p>
                                    <p className="text-xs text-slate-400">Run stages concurrently when possible</p>
                                </div>
                            </div>
                            <label className="flex items-center cursor-pointer" onClick={(e) => e.stopPropagation()}>
                                <input type="checkbox" className="sr-only" name="pipelineParallel" checked={pipeline.parallel}
                                    onChange={(e) => setPipeline({ ...pipeline, parallel: e.target.checked })} />
                                <div className="toggle-bg"></div>
                            </label>
                        </div>
                    </div>

                    {/* Pipeline stages */}
                    <div className="space-y-4">
                        <div className="flex justify-between items-center">
                            <div className="flex items-center space-x-2">
                                <h3 className="text-lg font-semibold text-slate-200">Pipeline Stages</h3>
                                <span className="px-2 py-1 text-xs bg-blue-600/20 text-blue-400 rounded-lg">{pipeline.stages.length}</span>
                            </div>
                            <button onClick={addStage}
                                className="flex items-center space-x-2 px-4 py-2 bg-gradient-to-r from-emerald-600 to-green-600
                                    hover:from-emerald-500 hover:to-green-500 text-white rounded-xl font-medium transition-all
                                    duration-200 shadow-lg hover:shadow-emerald-500/25">
                                <AddIcon className="w-4 h-4" />
                                <span>Add Stage</span>
                            </button>
                        </div>

                        {
                            pipeline.stages.map((s, i) => (
                                <div key={i} className="p-4 bg-gradient-to-r from-slate-800/50 to-slate-700/50 border border-slate-600/30 rounded-2xl
                                    relative hover:border-slate-500/50 transition-all duration-200">

                                    <div className="flex justify-between items-start mb-4">
                                        <div className="flex items-center space-x-3">
                                            <div className="w-8 h-8 rounded-xl bg-gradient-to-r from-blue-600 to-indigo-600 flex items-center
                                                justify-center text-white font-semibold text-sm">
                                                {i + 1}
                                            </div>
                                            <span className="text-sm font-medium text-slate-300">Stage {i + 1}</span>
                                        </div>
                                        <button onClick={(e) => deleteStage(e, i)} type="button"
                                            className="w-8 h-8 rounded-xl bg-red-500/20 hover:bg-red-500/30 text-red-400 hover:text-red-300
                                                transition-all duration-200 flex items-center justify-center">
                                            <DeleteIcon className="w-4 h-4" />
                                        </button>
                                    </div>

                                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                                        <div className="space-y-2">
                                            <label htmlFor={"stageName" + i} className="block text-sm font-medium text-slate-300">Stage Name</label>
                                            <input type="text" placeholder="e.g., build, test, deploy" name={"stageName" + i} value={s.name}
                                                onChange={(e) => updateStage(i, e.target.value, "name")}
                                                className="w-full px-3 py-2 bg-slate-700/50 border border-slate-600/50 rounded-lg text-slate-100
                                                    placeholder-slate-400 text-sm" />
                                        </div>
                                        <div className="space-y-2">
                                            <label htmlFor={"stageTask" + i} className="block text-sm font-medium text-slate-300">
                                                Command/Task
                                            </label>
                                            <input type="text" placeholder="npm run build" name={"stageTask" + i} value={s.task}
                                                onChange={(e) => updateStage(i, e.target.value, "task")}
                                                className="w-full px-3 py-2 bg-slate-700/50 border border-slate-600/50 rounded-lg text-slate-100
                                                    placeholder-slate-400 text-sm font-mono" />
                                        </div>

                                        <div className="space-y-2 lg:col-span-2">
                                            <div className="flex justify-between items-center">
                                                <label className="block text-sm font-medium text-slate-300">Command Arguments</label>
                                                <button onClick={(e) => addArg(e, i)} type="button"
                                                    className="flex items-center space-x-1 px-3 py-1 bg-gradient-to-r from-cyan-600 to-blue-600
                                                        hover:from-cyan-500 hover:to-blue-500 text-white rounded-lg text-xs font-medium
                                                        transition-all duration-200 shadow-sm hover:shadow-cyan-500/25">
                                                    <AddIcon className="w-3 h-3" />
                                                    <span>Add Arg</span>
                                                </button>
                                            </div>
                                            <div className="space-y-2">
                                                {(s.args || []).map((arg, argIndex) => (
                                                    <div key={`stage-${i}-arg-${argIndex}`} className="flex items-center space-x-2">
                                                        <div className="flex-1">
                                                            <input type="text" value={arg} name={`stageArg${i}_${argIndex}`}
                                                                placeholder={`Argument ${argIndex + 1} (e.g., --verbose, --output, ./dist)`}
                                                                onChange={(e) => updateArg(i, argIndex, e.target.value)}
                                                                className="w-full px-3 py-2 bg-slate-700/50 border border-slate-600/50 rounded-lg
                                                                    text-slate-100 placeholder-slate-400 text-sm font-mono" />
                                                        </div>
                                                        <button onClick={(e) => removeArg(e, i, argIndex)} type="button" className="w-8 h-8
                                                            rounded-lg bg-red-500/20 hover:bg-red-500/30 text-red-400 hover:text-red-300
                                                            transition-all duration-200 flex items-center justify-center flex-shrink-0">
                                                            <DeleteIcon className="w-4 h-4" />
                                                        </button>
                                                    </div>
                                                ))}
                                                {(!s.args || s.args.length === 0) && <ArgumentsEmptyState />}
                                            </div>
                                        </div>
                                        <div className="space-y-2">
                                            <label htmlFor={"stagePwd" + i} className="block text-sm font-medium text-slate-300">
                                                Working Directory
                                            </label>
                                            <input type="text" placeholder="/app/src" name={"stagePwd" + i} value={s.pwd}
                                                onChange={(e) => updateStage(i, e.target.value, "pwd")}
                                                className="w-full px-3 py-2 bg-slate-700/50 border border-slate-600/50 rounded-lg
                                                    text-slate-100 placeholder-slate-400 text-sm font-mono" />
                                        </div>

                                        <div className="flex items-center justify-between p-3 bg-slate-700/30 rounded-lg
                                            border border-slate-600/30">
                                            <div className="flex items-center space-x-3">
                                                <div className="w-8 h-8 rounded-lg bg-gradient-to-r from-orange-600 to-red-600 flex
                                                    items-center justify-center">
                                                    <SkipIcon className="w-4 h-4 text-white" />
                                                </div>
                                                <div>
                                                    <p className="font-medium text-slate-200 text-sm">Skip Stage</p>
                                                    <p className="text-xs text-slate-400">Skip this stage during pipeline execution</p>
                                                </div>
                                            </div>
                                            <label className="flex items-center cursor-pointer" onClick={(e) => e.stopPropagation()}>
                                                <input type="checkbox" className="sr-only" checked={s.skip || false} name={"stageSkip" + i}
                                                    onChange={(e) => updateStage(i, e.target.checked, "skip")} />
                                                <div className="toggle-bg"></div>
                                            </label>
                                        </div>

                                        <div className="space-y-2 lg:col-span-2">
                                            <div className="flex justify-between items-center">
                                                <label className="block text-sm font-medium text-slate-300">Environment Variables</label>
                                                <button onClick={(e) => addEnvVar(e, i)} type="button"
                                                    className="flex items-center space-x-1 px-3 py-1 bg-gradient-to-r from-green-600 to-emerald-600
                                                        hover:from-green-500 hover:to-emerald-500 text-white rounded-lg text-xs font-medium
                                                        transition-all duration-200 shadow-sm hover:shadow-green-500/25">
                                                    <AddIcon className="w-3 h-3" />
                                                    <span>Add Env Var</span>
                                                </button>
                                            </div>
                                            <div className="space-y-2">
                                                {(s.env || []).map((envVar, envIndex) => (
                                                    <div key={`stage-${i}-env-${envIndex}`} className="flex items-center space-x-2">
                                                        <div className="flex-1">
                                                            <input type="text" value={envVar} name={`stageEnv${i}_${envIndex}`}
                                                                placeholder="Environment variable (e.g., NODE_ENV=production, DEBUG=true)"
                                                                onChange={(e) => updateEnvVar(i, envIndex, e.target.value)}
                                                                className="w-full px-3 py-2 bg-slate-700/50 border border-slate-600/50 rounded-lg
                                                                    text-slate-100 placeholder-slate-400 text-sm font-mono" />
                                                        </div>
                                                        <button onClick={(e) => removeEnvVar(e, i, envIndex)} type="button"
                                                            className="w-8 h-8 rounded-lg bg-red-500/20 hover:bg-red-500/30 text-red-400
                                                                hover:text-red-300 transition-all duration-200 flex items-center
                                                                justify-center flex-shrink-0">
                                                            <DeleteIcon className="w-4 h-4" />
                                                        </button>
                                                    </div>
                                                ))}
                                                {(!s.env || s.env.length === 0) && (
                                                    <EnvVarsEmptyState />
                                                )}
                                            </div>
                                        </div>

                                        <div className="space-y-2 lg:col-span-2">
                                            <label htmlFor={"stageDepend" + i} className="block text-sm font-medium text-slate-300">
                                                Dependencies
                                            </label>
                                            <div className="space-y-3">
                                                <div className="flex flex-wrap gap-2">
                                                    {s.depends_on.map((d, j) => (
                                                        <span key={s.name + "dep" + j}
                                                            className="inline-flex items-center px-3 py-1 bg-gradient-to-r from-sky-600 to-blue-600
                                                                text-white text-md rounded-lg">
                                                            <span>{d}</span>
                                                            <button type="button" className="ml-2 text-sky-200 hover:text-white transition-colors"
                                                                onClick={(e) => removeDependency(e, i, j)}>
                                                                <CloseIcon className="w-4 h-4" />
                                                            </button>
                                                        </span>
                                                    ))}
                                                </div>
                                                <select onChange={(e) => addDependency(e, i)} name={"stageDepend" + i}
                                                    className="w-full px-3 py-2 bg-slate-700/50 border border-slate-600/50 rounded-lg
                                                        text-slate-100 text-sm">
                                                    <option value="">+ Add dependency on previous stage</option>
                                                    {pipeline.stages.slice(0, i)
                                                        .filter(dStages => !pipeline.stages[i].depends_on.includes(dStages.name)
                                                            && dStages.name.trim().length > 0)
                                                        .map((dStages, j) => <option key={j + i} value={dStages.name}>{dStages.name}</option>)}
                                                </select>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            ))
                        }
                    </div>

                    {/* Pipeline vars */}
                    <div className="space-y-4">
                        <div className="flex justify-between items-center">
                            <div className="flex items-center space-x-2">
                                <h3 className="text-lg font-semibold text-slate-200">Pipeline Variables</h3>
                                <span className="px-2 py-1 text-xs bg-purple-600/20 text-purple-400 rounded-lg">{pipeline.variables.length}</span>
                            </div>
                            <button onClick={addVariable}
                                className="flex items-center space-x-2 px-4 py-2 bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500
                                    hover:to-indigo-500 text-white rounded-xl font-medium transition-all duration-200
                                    shadow-lg hover:shadow-purple-500/25">
                                <AddIcon className="w-4 h-4" />
                                <span>Add Variable</span>
                            </button>
                        </div>

                        <div className="space-y-3"> {
                            pipeline.variables.map((v, i) => (
                                <div key={i} className="grid grid-cols-1 lg:grid-cols-2 gap-4 p-4 bg-gradient-to-r from-slate-800/30 to-slate-700/30
                                    border border-slate-600/30 rounded-xl">
                                    <div className="space-y-2">
                                        <label className="block text-sm font-medium text-slate-300">Variable Name</label>
                                        <input type="text" placeholder="API_KEY" name={"varKey" + i} value={v[0] ?? ""}
                                            onChange={(e) => updateVariable(i, e.target.value, true)}
                                            className="w-full px-3 py-2 bg-slate-700/50 border border-slate-600/50 rounded-lg text-slate-100
                                                placeholder-slate-400 focus:ring-2 focus:ring-purple-500/50 focus:border-purple-500/50
                                                transition-all duration-200 text-sm font-mono" />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="block text-sm font-medium text-slate-300">Variable Value</label>
                                        <input type="text" placeholder="your-api-key-here" name={"varValue" + i} value={v[1] ?? ""}
                                            onChange={(e) => updateVariable(i, e.target.value, false)}
                                            className="w-full px-3 py-2 bg-slate-700/50 border border-slate-600/50 rounded-lg text-slate-100
                                                placeholder-slate-400 focus:ring-2 focus:ring-purple-500/50 focus:border-purple-500/50
                                                transition-all duration-200 text-sm font-mono" />
                                    </div>
                                </div>
                            ))
                        }
                        </div>
                    </div>
                </div>

                {/* Action Buttons */}
                <div className="flex justify-between items-center pt-6 border-t border-slate-600/30">
                    <div className="flex space-x-3">{
                        existing ? <button onClick={updatePipeline}
                            className="flex items-center space-x-2 px-6 py-3 bg-gradient-to-r from-blue-600 to-sky-600 hover:from-blue-500
                                hover:to-sky-500 text-white rounded-xl font-medium transition-all duration-200 shadow-lg hover:shadow-blue-500/25">
                            <UploadIcon className="w-4 h-4" />
                            <span>Update Pipeline</span>
                        </button> :
                            <button onClick={addPipeline}
                                className="flex items-center space-x-2 px-6 py-3 bg-gradient-to-r from-green-600 to-emerald-600
                                    hover:from-green-500 hover:to-emerald-500 text-white rounded-xl font-medium transition-all duration-200
                                    shadow-lg hover:shadow-green-500/25">
                                <AddIcon className="w-4 h-4" />
                                <span>Create Pipeline</span>
                            </button>
                    }
                    <button onClick={() => setShowPipelineModal(false)}
                        className="flex items-center space-x-2 px-6 py-3 bg-slate-700/50 hover:bg-slate-600/50 border border-slate-600/50
                            hover:border-slate-500/50 text-slate-300 hover:text-slate-200 rounded-xl font-medium transition-all duration-200">
                        <CloseIcon className="w-4 h-4" />
                        <span>Cancel</span>
                    </button>
                    </div>
                    {existing && (
                        <button onClick={(e) => deletePipeline(e, pipeline.name)} type="button"
                            className="flex items-center space-x-2 px-6 py-3 bg-gradient-to-r from-red-700 to-rose-700 hover:from-red-600
                                hover:to-rose-600 text-white rounded-xl font-medium transition-all duration-200 shadow-lg hover:shadow-red-500/25">
                            <DeleteIcon className="w-4 h-4" />
                            <span>Delete Pipeline</span>
                        </button>
                    )}
                </div>
            </form>
        </div>;
    };

    return (
        <>
            {
                showPipelineModal ? <Modal close={() => setShowPipelineModal(false)}>
                    { addPipelineTemplate() }
                </Modal>: <></>
            }

            <div className="w-full lg:w-1/3 h-screen overflow-y-hidden bg-gradient-to-b from-slate-900 to-slate-800 border-r
                border-slate-700/50 flex flex-col backdrop-blur-sm">
                {/* Search filter */}
                <div className="relative mx-4 mt-4 mb-2">
                    <input type="text" placeholder="Search pipelines..." value={searchFilter} onChange={(e) => setSearchFilter(e.target.value)}
                        className="py-3 px-4 pl-11 block w-full shadow-lg text-sm outline-none text-slate-200 bg-slate-800/60 backdrop-blur-sm border
                            border-slate-600/50 rounded-xl focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500/50 transition-all
                            duration-200 placeholder-slate-400"/>
                    <div className="absolute inset-y-0 left-0 flex items-center pointer-events-none pl-4">
                        <SearchIcon className="h-4 w-4 text-slate-400" />
                    </div>
                    {searchFilter && (
                        <button onClick={() => setSearchFilter("")}
                            className="absolute inset-y-0 right-0 flex items-center pr-4 text-slate-400 hover:text-slate-200 transition-colors">
                            <CloseIcon className="w-4 h-4" />
                        </button>
                    )}
                </div>

                {/* Pipelines header */}
                <div className="w-full px-6 py-4 border-b border-slate-700/50">
                    <div className="flex justify-between items-center mb-3">
                        <div className="flex items-center space-x-2">
                            <h2 className="text-lg font-semibold text-slate-200">Pipelines</h2>
                            <button onClick={getPipelines} title="Refresh pipelines"
                                className="p-1.5 rounded-lg bg-slate-700/50 hover:bg-slate-600/50 text-slate-400 hover:text-slate-200">
                                <RefreshIcon className="w-4 h-4" />
                            </button>
                        </div>
                        <button onClick={showAddPipelineDialog} title="Add new pipeline"
                            className="w-9 h-9 rounded-full bg-gradient-to-r from-emerald-600 to-emerald-500 hover:from-emerald-500
                                hover:to-emerald-400 shadow-lg hover:shadow-emerald-500/25 flex items-center
                                justify-center text-white font-semibold">
                            <AddIcon className="w-4 h-4" />
                        </button>
                    </div>

                    {/* Pipeline Stats */}
                    {pipelines && pipelines.length > 0 && (
                        <div className="grid grid-cols-3 gap-2 text-xs">
                            <div className="px-3 py-2 bg-slate-800/50 rounded-lg border border-slate-700/30 text-center">
                                <div className="font-semibold text-slate-200">{pipelines.length}</div>
                                <div className="text-slate-400">Total</div>
                            </div>
                            <div className="px-3 py-2 bg-slate-800/50 rounded-lg border border-slate-700/30 text-center">
                                <div className="font-semibold text-emerald-400">{pipelines.filter(p => p.status === "running").length}</div>
                                <div className="text-slate-400">Running</div>
                            </div>
                            <div className="px-3 py-2 bg-slate-800/50 rounded-lg border border-slate-700/30 text-center">
                                <div className="font-semibold text-red-400">{pipelines.filter(p => p.status === "failed").length}</div>
                                <div className="text-slate-400">Failed</div>
                            </div>
                        </div>
                    )}
                </div>

                {/* Pipeline list */}
                <div className="w-full overflow-y-auto flex flex-col scroll-smooth py-2" ref={pipelinesContainer}>{
                    pipelines?.filter(pipelineFilter).length === 0 ? (
                        <PipelineEmptyState
                            searchFilter={searchFilter}
                            onCreatePipeline={showAddPipelineDialog}
                        />
                    ) : (
                        pipelines?.filter(pipelineFilter).sort(pipelineSorter).map((record, key) => (
                            <div key={key} onClick={() => showDetails(record.name)} className="mx-3 mb-2 p-4 rounded-xl bg-slate-800/40
                                backdrop-blur-sm border border-slate-700/30 hover:bg-slate-700/60 hover:border-slate-600/50 cursor-pointer
                                hover:shadow-lg hover:shadow-slate-900/20">
                                <div className="flex justify-between items-start mb-3">
                                    <div className="flex items-center space-x-3">
                                        <div className="relative">
                                            <div title={record.status} className={"w-3 h-3 rounded-full border-0 " + record.status}></div>
                                            {
                                                record.status === "running" && (
                                                    <div className="absolute inset-0 w-3 h-3 rounded-full animate-ping opacity-75 running">
                                                    </div>
                                                )
                                            }
                                        </div>
                                        <div className="flex-1 min-w-0">
                                            <div className="flex items-center space-x-2">
                                                <h3 className="font-semibold text-slate-200 transition-colors duration-200 truncate">
                                                    {record.name}
                                                </h3>
                                                {record.status === "running" && (
                                                    <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium
                                                        bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                                                        <div className="w-1.5 h-1.5 running rounded-full mr-1 animate-pulse"></div>
                                                        Running
                                                    </span>
                                                )}
                                                {!!record.last_run && record.last_run > Date.now() - 24 * 60 * 60 * 1000 && (
                                                    <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium
                                                        bg-blue-500/20 text-blue-400 border border-blue-500/30">Recent</span>
                                                )}
                                            </div>
                                        </div>
                                    </div>

                                    {/* Quick Actions */}
                                    <div className="flex items-center space-x-1">
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                // Add run pipeline functionality
                                                console.log("Run pipeline:", record.name);
                                            }}
                                            title="Run Pipeline" className="p-1.5 rounded-lg hover:bg-emerald-600/20 text-slate-400
                                                hover:text-emerald-400 transition-all duration-200">
                                            <PlayIcon className="w-4 h-4" />
                                        </button>
                                        <button onClick={(e) => showEditPipelineDialog(e, record.name)} title="Edit Pipeline"
                                            className="p-1.5 rounded-lg hover:bg-slate-600/50 text-slate-400 hover:text-slate-200
                                                transition-all duration-200">
                                            <EditIcon className="w-4 h-4" />
                                        </button>
                                    </div>
                                </div>

                                {/* Pipeline Metrics */}
                                <div className="grid grid-cols-1 gap-3 text-xs">
                                    <div className="flex items-center space-x-2 px-3 py-2 bg-slate-700/30 rounded-lg border border-slate-600/30">
                                        <ClockIcon className="w-3.5 h-3.5 text-slate-400" />
                                        <span className="text-slate-400">Last run</span>
                                        <span className="text-slate-300 font-medium ml-auto">{formatLastRun(record.last_run)}</span>
                                    </div>
                                </div>

                                {/* Pipeline Stages Preview & Additional Info */}
                                <div className="mt-3 pt-3 border-t border-slate-700/50 space-y-2">
                                    {record.stages && record.stages.length > 0 && (
                                        <>
                                            {record.status === "complete" && (
                                                <div className="flex items-center space-x-1 text-xs text-emerald-400">
                                                    <CheckIcon className="w-3 h-3" />
                                                    <span>Completed</span>
                                                </div>
                                            )}
                                            {record.status === "failed" && (
                                                <div className="flex items-center space-x-1 text-xs text-red-400">
                                                    <CloseIcon className="w-3 h-3" />
                                                    <span>Failed</span>
                                                </div>
                                            )}
                                        </>
                                    )}
                                </div>
                            </div>
                        ))
                    )}
                </div>
            </div>
        </>
    );
};

export default PipelineListPanel;