import { useState, useEffect, useRef } from "react";
import { request } from "../utils/Fetch";

import { useAppContext } from "../AppContext";
import { NotificationType, useNotificationContext } from "./Notification";

import Modal from "./Modal";

const stageInitialState = {
    name: "",
    task: "",
    depends_on: [],
    pwd: ""
};

const pipelineInitialState = {
    name: "",
    parallel: false,
    stages: [JSON.parse(JSON.stringify(stageInitialState))],
    variables: [[]]
};

const PipelineListPanel = () => {
    const { display: displayNotification } = useNotificationContext();
    const { pipelines, setPipelines, setSelectedPipeline, setShowDetails } = useAppContext();
    const [searchFilter, setSearchFilter] = useState("");
    const [showAddModal, setShowAddModal] = useState(false);
    const [pipeline, setPipeline] = useState(pipelineInitialState);

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
        console.log(newStages);
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

    const showAddPipelineDialog = () => setShowAddModal(true);
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
                    // TODO: trigger refresh to get new pipelines
                    setShowAddModal(false);
                    displayNotification({ message: "Pipeline registered", type: NotificationType.Success });
                    setPipeline({ ...pipelineInitialState });
                    getPipelines();
                } else {
                    displayNotification({ message: "An error occurred adding pipeline: " + msg, type: NotificationType.Error });
                }
            }
        });
    };

    const streamSorter = (a, b) => {
        // TODO: update with sorting priorities
        return a.name?.localeCompare(b.name);
    };

    const streamFilter = ({ name }) => searchFilter === "" || name.toLowerCase().includes(searchFilter.toLowerCase());

    const addPipelineTemplate = () => {
        return <div className="w-full h-full lg:w-2/5 lg:h-4/5 mx-auto shadow-sm bg-stone-800 md:rounded-md flex flex-col">
            <div className="w-full p-4 flex justify-between border-b-1 border-stone-700">
                <h2 className="text-2xl font-bold">Add Pipeline</h2>
                <button onClick={() => setShowAddModal(false)}
                    className="w-8 h-8 rounded-full bg-red-900 hover:bg-red-700 focus:bg-red-700 cursor-pointer">&times;</button>
            </div>
            <form className="flex flex-col justify-between p-2 lg:p-4 h-full overflow-y-scroll">
                <div>
                    <div className="my-2">
                        <label htmlFor="pipelineName" className="block font-bold mb-2">Pipeline Name</label>
                        <input type="text" placeholder="enter pipeline name" name="pipelineName" value={pipeline.name}
                            onChange={(e) => setPipeline({ ...pipeline, name: e.target.value })}
                            className="shadow border-1 border-stone-300 rounded-sm w-full py-2 px-3" />
                    </div>
                    <div className="my-6">
                        <label htmlFor="pipelineParallel" className="block font-bold mb-2">Run Parallel</label>
                        <label className="flex items-center cursor-pointer relative mb-4" onClick={(e) => e.stopPropagation()}>
                            <input type="checkbox" className="sr-only" name="pipelineParallel" checked={pipeline.parallel}
                                onChange={(e) => setPipeline({ ...pipeline, parallel: e.target.checked })} />
                            <div className="toggle-bg bg-gray-400 border-2 border-gray-400 h-6 w-11 rounded-full"></div>
                        </label>
                    </div>
                    <div className="my-2">
                        <div className="flex justify-between">
                            <label className="block font-bold mb-4 mr-4">Stages</label>
                            <button onClick={addStage}
                                className="w-8 h-8 rounded-full bg-green-800 hover:bg-green-600 focus:bg-green-600 cursor-pointer">+</button>
                        </div>
                        {
                            pipeline.stages.map((s, i) => (
                                <div className="mb-4 flex flex-col border-1 border-stone-700 p-4 pt-2 rounded-md relative" key={i}>
                                    <button onClick={(e) => deleteStage(e, i)}
                                        className="w-8 h-8 rounded-full bg-red-900 hover:bg-red-700 focus:bg-red-700
                                            cursor-pointer absolute top-2 right-2">&times;</button>
                                    <div className="my-2">
                                        <label htmlFor={"stageName" + i} className="block font-bold mb-2">Name</label>
                                        <input type="text" placeholder="enter stage name" name={"stageName" + i} value={s.name}
                                            onChange={(e) => updateStage(i, e.target.value, "name")}
                                            className="shadow border-1 border-stone-300 rounded-sm w-full
                                                    lg:w-1/2 py-2 px-3 mr-0 lg:mr-2 mb-2 lg:mb-0" />
                                    </div>
                                    <div className="my-2">
                                        <label htmlFor={"stageTask" + i} className="block font-bold mb-2">Task</label>
                                        <input type="text" placeholder="enter stage task" name={"stageTask" + i} value={s.task}
                                            onChange={(e) => updateStage(i, e.target.value, "task")}
                                            className="shadow border-1 border-stone-300 rounded-sm w-full py-2 px-3" />
                                    </div>
                                    <div className="my-2">
                                        <label htmlFor={"stagePwd" + i} className="block font-bold mb-2">Pwd</label>
                                        <input type="text" placeholder="enter stage task" name={"stagePwd" + i} value={s.pwd}
                                            onChange={(e) => updateStage(i, e.target.value, "pwd")}
                                            className="shadow border-1 border-stone-300 rounded-sm w-full py-2 px-3" />
                                    </div>
                                    <div className="my-2">
                                        <label htmlFor={"stageDepend" + i} className="block font-bold mb-2">Dependencies</label>
                                        <div>
                                            {s.depends_on.map((d, j) => (
                                                <span key={s.name + "dep" + j}
                                                    className="text-lg text-gray-900 bg-sky-600 rounded-md p-1 pl-4 m-1 inline-block text-nowrap">
                                                    <span>{d}</span>
                                                    <span className="text-2xl text-gray-800 hover:text-red-500 cursor-pointer ml-2"
                                                        onClick={(e) => removeDependency(e, i, j)}>&times;</span>
                                                </span>
                                            ))}
                                        </div>
                                        <select onChange={(e) => addDependency(e, i)} name={"stageDepend" + i}
                                            className="mt-4 bg-stone-700 text-stone-200">
                                            <option value="">+ Add Dependency</option>
                                            {pipeline.stages.slice(0, i)
                                                .filter(dStages => !pipeline.stages[i].depends_on.includes(dStages.name)
                                                    && dStages.name.trim().length > 0)
                                                .map((dStages, j) => <option key={j + i} value={dStages.name}>{dStages.name}</option>)}
                                        </select>
                                    </div>
                                </div>
                            ))
                        }
                    </div>
                    <div className="my-2">
                        <div className="flex justify-between">
                            <label className="block font-bold mb-4">Variables</label>
                            <button onClick={addVariable}
                                className="w-8 h-8 rounded-full bg-green-800 hover:bg-green-600 focus:bg-green-600 cursor-pointer">
                                +</button>
                        </div>
                        {
                            pipeline.variables.map((v, i) => (
                                <div className="mb-8 lg:mb-4 flex flex-col lg:flex-row" key={i}>
                                    <input type="text" placeholder="enter variable name" name={"varKey" + i} value={v[0]}
                                        onChange={(e) => updateVariable(i, e.target.value, true)}
                                        className="shadow border-1 border-stone-300 rounded-sm w-full lg:w-1/2 py-2
                                                px-3 mr-0 lg:mr-2 mb-2 lg:mb-0" />
                                    <input type="text" placeholder="enter variable value" name={"varValue" + i} value={v[1]}
                                        onChange={(e) => updateVariable(i, e.target.value, false)}
                                        className="shadow border-1 border-stone-300 rounded-sm w-full lg:w-1/2 py-2 px-3" />
                                </div>
                            ))
                        }
                    </div>
                </div>
                <div className="flex w-fit space-x-2">
                    <button onClick={addPipeline}
                        className="rounded-sm border-1 hover:bg-green-700 focus:bg-green-700 block ml-auto
                                cursor-pointer px-4 py-2 font-bold">
                        Add</button>
                    {/* TODO: should this clear the form */}
                    <button onClick={() => setShowAddModal(false)}
                        className="rounded-sm bg-red-900 hover:bg-red-700 focus:bg-red-700 block ml-auto
                                cursor-pointer px-4 py-2 font-bold">
                        Cancel</button>
                </div>
            </form>
        </div>;
    };

    return (
        <div className="w-full lg:w-1/3 h-screen overflow-y-hidden bg-indigo-900 flex flex-col">
            {/* add pipeline modal */}
            {
                showAddModal ? <Modal close={() => setShowAddModal(false)}>
                    { addPipelineTemplate() }
                </Modal>: <></>
            }

            {/* Search filter */}
            <div className="relative flex rounded-md shadow-sm">
                <input type="text" placeholder="search filter" value={searchFilter} onChange={(e) => setSearchFilter(e.target.value)}
                    className="py-3 px-4 pl-11 block w-full shadow-sm text-base outline-none text-gray-300 bg-stone-800"/>
                <div className="absolute inset-y-0 left-0 flex items-center pointer-events-none pl-4">
                    <svg className="h-4 w-4 text-gray-400" xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                        fill="currentColor" viewBox="0 0 16 16">
                        <path d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1
                            1 0 0 0 1.415-1.414l-3.85-3.85a1.007 1.007 0 0 0-.115-.1zM12 6.5a5.5 5.5 0 1 1-11 0 5.5
                            5.5 0 0 1 11 0z"/>
                    </svg>
                </div>
            </div>

            {/* Pipelines header */}
            <div className="w-full px-2 py-4 border-b-1 border-r-1 border-t-1 border-indigo-700 flex justify-between">
                <label className="block font-bold">Pipelines</label>
                <button onClick={showAddPipelineDialog} className="w-8 h-8 rounded-full bg-green-700 hover:bg-green-600
                    focus:bg-green-500 block ml-auto cursor-pointer">+</button>
            </div>

            {/* Pipeline list */}
            <div className="w-full overflow-y-auto flex flex-col scroll-smooth border-r-1 border-indigo-700" ref={pipelinesContainer}>
                { pipelines?.filter(streamFilter).sort(streamSorter).map((record, key) => (
                    <div key={key} onClick={() => showDetails(record.name)} className="hover:bg-blue-600 flex
                        justify-between px-2 items-center hover:cursor-pointer">
                        <div className="flex items-center">
                            <div title={record.status}
                                className={"mr-2 w-6 h-6 rounded-full border-2 border-stone-100 " + record.status}>
                            </div>
                            <span className="py-2 px-1 font-bold">{record.name}</span>
                        </div>
                        <div className="flex flex-col py-2 px-1 font-light text-indigo-300">
                            <span>{record.last_run ?? "-"}</span>
                            <span>{record.runtime ?? "--:--"}</span>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default PipelineListPanel;