document.addEventListener("DOMContentLoaded",()=>{
  const grid=document.getElementById("grid");
  const newBtn=document.getElementById("new-btn");
  const diffSel=document.getElementById("diff-select");
  const btnSolve=document.getElementById("solve");
  const btnValidate=document.getElementById("validate");
  const btnHint=document.getElementById("hint");
  const btnSave=document.getElementById("save");
  const btnLoad=document.getElementById("load");
  const nameInput=document.getElementById("name-input");
  const notesInput=document.getElementById("notes-input");
  const autoCand=document.getElementById("auto-candidates");

  // --- state ---
  let sel={r:0,c:0};
  let notesMode=false; // manual notes mode toggle (press 'n')

  // Undo/Redo + autosave state
  const MAX_STACK=100;
  const undoStack=[];
  const redoStack=[];
  let currentId = localStorage.getItem("sudoku.currentId") || "";
  let dirty=false;
  let autosaveTimer=null;
  const AUTOSAVE_DEBOUNCE_MS=2000;
  const AUTOSAVE_INTERVAL_MS=10000;

  // --- helpers to build cells ---
  function mkCell(r,c){
    const d=document.createElement("div");
    d.className="cell"; d.setAttribute("role","gridcell");
    d.dataset.r=r; d.dataset.c=c; d.tabIndex=0;
    d.dataset.val="0";        // 0 means empty
    d.dataset.notes="";       // manual notes like '135'
    d.addEventListener("click",()=>{sel={r,c}; highlight();});
    return d;
  }

  // build grid
  for(let r=0;r<9;r++){ for(let c=0;c<9;c++){ grid.appendChild(mkCell(r,c)); } }

  function idx(r,c){return r*9+c;}
  function cell(r,c){return grid.children[idx(r,c)];}
  function getVal(r,c){return parseInt(cell(r,c).dataset.val||"0",10)||0;}
  function setFixed(r,c,fx){cell(r,c).classList.toggle("fixed",!!fx);}
  function setVal(r,c,v){
    const el=cell(r,c);
    if(el.classList.contains("fixed")) return;
    pushUndo(); // capture state before mutation
    const n = Math.max(0, Math.min(9, v|0));
    el.dataset.val=String(n);
    if(n===0){ renderCell(r,c); } else { el.dataset.notes=""; renderCell(r,c); }
    markDirty();
  }
  function toggleNote(r,c,d){
    const el=cell(r,c);
    if(el.classList.contains("fixed")) return;
    pushUndo();
    let s=el.dataset.notes||"";
    if(s.includes(d)){ s=s.replace(d,""); } else { s=[...s,d].sort().join(""); }
    el.dataset.notes=s;
    if(getVal(r,c)===0 && !autoCand.checked){ renderCell(r,c); }
    markDirty();
  }
  function renderCell(r,c){
    const el=cell(r,c);
    el.textContent="";
    el.style.outline="";
    const v=getVal(r,c);
    if(v>0){
      el.textContent=String(v);
      return;
    }
    // show overlay (manual notes or auto-candidates)
    const overlay=document.createElement("div");
    overlay.style.fontSize="10px";
    overlay.style.lineHeight="12px";
    overlay.style.textAlign="center";
    overlay.style.whiteSpace="pre-wrap";
    overlay.style.opacity="0.85";
    overlay.style.userSelect="none";
    const toShow = autoCand.checked ? candidatesString(r,c) : (cell(r,c).dataset.notes||"");
    overlay.textContent = toShow.split("").join(" ");
    el.appendChild(overlay);
  }
  function renderAll(){
    for(let r=0;r<9;r++){ for(let c=0;c<9;c++){ renderCell(r,c); } }
    highlight();
  }

  function highlight(){ [...grid.children].forEach(el=>el.classList.remove("sel")); cell(sel.r,sel.c).classList.add("sel"); }
  function move(dr,dc){ sel.r=(sel.r+dr+9)%9; sel.c=(sel.c+dc+9)%9; highlight(); }
  function clearClass(cls){ [...grid.children].forEach(el=>el.classList.remove(cls)); }
  function clearInlineOutlines(){ [...grid.children].forEach(el=>el.style.outline=""); }

  // --- board I/O ---
  function getBoard(){
    const a=[...Array(9)].map(()=>Array(9).fill(0));
    for(let r=0;r<9;r++){ for(let c=0;c<9;c++){ a[r][c]=getVal(r,c); } }
    return a;
  }
  function getFixed(){
    const f=[...Array(9)].map(()=>Array(9).fill(false));
    for(let r=0;r<9;r++){ for(let c=0;c<9;c++){ f[r][c]=cell(r,c).classList.contains("fixed"); } }
    return f;
  }
  function getNotes(){
    const n=[...Array(9)].map(()=>Array(9).fill(""));
    for(let r=0;r<9;r++){ for(let c=0;c<9;c++){ n[r][c]=cell(r,c).dataset.notes||""; } }
    return n;
  }
  function setNotes(notes){
    if(!notes) return;
    for(let r=0;r<9;r++){ for(let c=0;c<9;c++){
      cell(r,c).dataset.notes = (notes?.[r]?.[c]||"");
      if(getVal(r,c)===0 && !autoCand.checked){ renderCell(r,c); }
    } }
  }
  function setBoard(board,fixed){
    for(let r=0;r<9;r++){ for(let c=0;c<9;c++){
      const v = board?.[r]?.[c]||0;
      const fx = !!(fixed?.[r]?.[c]);
      cell(r,c).dataset.val=String(v);
      cell(r,c).dataset.notes="";
      setFixed(r,c,fx);
    } }
    renderAll();
  }

  // --- constraints ---
  function allowedAt(r,c,v){
    // row/col
    for(let i=0;i<9;i++){ if(getVal(r,i)===v || getVal(i,c)===v) return false; }
    // box
    const br=Math.floor(r/3)*3, bc=Math.floor(c/3)*3;
    for(let dr=0;dr<3;dr++){ for(let dc=0;dc<3;dc++){ if(getVal(br+dr,bc+dc)===v) return false; } }
    return true;
  }
  function candidatesString(r,c){
    if(getVal(r,c)!==0) return "";
    let s="";
    for(let v=1;v<=9;v++){ if(allowedAt(r,c,v)) s+=String(v); }
    return s;
  }

  // --- Undo/Redo ---
  function snapshot(){
    return {
      board: getBoard(),
      fixed: getFixed(),
      notes: getNotes(),
      sel: {...sel},
      meta: {
        name: (nameInput?.value||"").trim(),
        notesText: (notesInput?.value||"").trim(),
        difficulty: (diffSel?.value||"medium")
      }
    };
  }
  function applySnapshot(s){
    if(!s) return;
    setBoard(s.board, s.fixed);
    setNotes(s.notes);
    if(s.sel) sel = {r:s.sel.r|0, c:s.sel.c|0};
    if(s.meta){
      nameInput.value = s.meta.name || "";
      notesInput.value = s.meta.notesText || "";
      if(s.meta.difficulty) diffSel.value = s.meta.difficulty;
    }
    renderAll();
    clearClass("conflict"); clearInlineOutlines();
  }
  function pushUndo(){
    // prevent pushing identical immediate states by checking last snapshot's board reference (cheap heuristic)
    const snap = snapshot();
    undoStack.push(snap);
    if(undoStack.length>MAX_STACK) undoStack.shift();
    // changing state invalidates redo
    redoStack.length = 0;
  }
  function undo(){
    if(undoStack.length===0) return;
    const current = snapshot();
    const state = undoStack.pop();
    redoStack.push(current);
    if(redoStack.length>MAX_STACK) redoStack.shift();
    applySnapshot(state);
    markDirty(); // content changed
  }
  function redo(){
    if(redoStack.length===0) return;
    const current = snapshot();
    const state = redoStack.pop();
    undoStack.push(current);
    if(undoStack.length>MAX_STACK) undoStack.shift();
    applySnapshot(state);
    markDirty();
  }

  // --- keyboard ---
  document.addEventListener("keydown",e=>{
    const k=e.key;
    // Undo/Redo shortcuts
    if(e.ctrlKey && !e.shiftKey && k==="z"){ e.preventDefault(); return undo(); }
    if((e.ctrlKey && k==="y") || (e.ctrlKey && e.shiftKey && k==="Z")){ e.preventDefault(); return redo(); }

    if(k==="ArrowUp"||k==="k") return move(-1,0);
    if(k==="ArrowDown"||k==="j") return move(1,0);
    if(k==="ArrowLeft"||k==="h") return move(0,-1);
    if(k==="ArrowRight"||k==="l") return move(0,1);
    if(k==="n"||k==="N"){ e.preventDefault(); notesMode=!notesMode; return; }
    if(k==="?" ){ e.preventDefault(); return doHint(); }
    if(k==="f"||k==="F"){ e.preventDefault(); return autoFillSingles(); }
    if(/^[1-9]$/.test(k)){
      const d=k;
      if(notesMode){ return toggleNote(sel.r,sel.c,d); }
      return setVal(sel.r,sel.c,parseInt(k,10));
    }
    if(k==="0"||k===" "||k==="Backspace"||k==="Delete"){ return setVal(sel.r,sel.c,0); }
  });

  // --- numpad ---
  for(const b of document.querySelectorAll(".numpad button")){
    b.addEventListener("click",()=>{
      const v=parseInt(b.dataset.num,10)||0;
      if(v===0) setVal(sel.r,sel.c,0); else if(notesMode) toggleNote(sel.r,sel.c,String(v)); else setVal(sel.r,sel.c,v);
    });
  }

  // --- API helpers ---
  async function api(path, payload){
    const res=await fetch(path,{
      method:"POST",
      headers:{"Content-Type":"application/json"},
      body: JSON.stringify(payload??{})
    });
    return res.json();
  }

  function markConflicts(conf){ clearClass("conflict"); if(!conf) return; for(const p of conf){ cell(p.row,p.col).classList.add("conflict"); } }
  function markHint(cells,msg){ clearInlineOutlines(); if(!cells||!cells.length){ alert("No simple hint found."); return; } for(const p of cells){ cell(p.row,p.col).style.outline="2px dashed orange"; } if(msg) console.log(msg); }

  // --- actions ---
  btnValidate?.addEventListener("click",async()=>{
    try{
      const data=await api("/api/validate",{board:getBoard()});
      markConflicts(data.conflicts);
      if(data.ok){ console.log("OK: no conflicts"); }
    }catch(e){ console.error("Validate failed",e); }
  });

  btnSolve?.addEventListener("click",async()=>{
    try{
      const data=await api("/api/solve",{board:getBoard()});
      if(data.board){ setBoard(data.board); clearClass("conflict"); clearInlineOutlines(); }
      else{ alert("Solve failed: "+(data.error||"unknown")); }
    }catch(e){ alert("Solve error: "+e); }
  });

  async function doGenerate(diff){
    try{
      const data=await api("/api/generate",{difficulty:diff});
      if(data.board){
        // Generating a new puzzle = new identity
        currentId="";
        localStorage.removeItem("sudoku.currentId");
        nameInput.value=""; notesInput.value="";
        setBoard(data.board.board, data.board.fixed);
        clearClass("conflict"); clearInlineOutlines();
        undoStack.length=0; redoStack.length=0;
        markDirty(); // fresh board is unsaved
      }
      else{ alert("Generate failed: "+(data.error||"unknown")); }
    }catch(e){ alert("Generate error: "+e); }
  }

  // single New button + last-used difficulty
  newBtn?.addEventListener("click",()=>{
    const d=diffSel.value||"medium";
    localStorage.setItem("sudoku.lastDiff", d);
    doGenerate(d);
  });
  // restore last used difficulty
  (function(){
    const d=localStorage.getItem("sudoku.lastDiff");
    if(d && [...diffSel.options].some(o=>o.value===d)){ diffSel.value=d; }
  })();

  btnHint?.addEventListener("click",()=>doHint());
  async function doHint(){
    try{
      const data=await api("/api/hint",{board:getBoard(),maxTier:"singles"});
      if(data.found && data.hint){ markHint(data.hint.cells, data.hint.message); }
      else { alert("No simple hint found."); clearInlineOutlines(); }
    }catch(e){ alert("Hint error: "+e); }
  }

  // Auto-fill singles action (hotkey F)
  function autoFillSingles(){
    let filled=0;
    for(let r=0;r<9;r++){ for(let c=0;c<9;c++){
      if(getVal(r,c)!==0) continue;
      const s=candidatesString(r,c);
      if(s.length===1){
        setVal(r,c, parseInt(s,10));
        filled++;
      }
    }};
    if(filled===0) console.log("No singles to auto-fill.");
    if(autoCand.checked) renderAll(); else { for(let r=0;r<9;r++){ for(let c=0;c<9;c++){ if(getVal(r,c)===0) renderCell(r,c); } } }
    return filled;
  }

  // Save helpers (manual + autosave)
  async function savePuzzle({silent}={silent:true}){
    try{
      const payload={
        id: currentId || "",
        name: (nameInput?.value||"").trim(),
        notes: (notesInput?.value||"").trim(),
        // omit difficulty to avoid server enum mismatch; storage defaults to 'medium' if absent
        board:{board:getBoard(), fixed:getFixed()}
      };
      const res=await fetch("/api/save",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(payload)});
      const data=await res.json();
      if(data.id){
        currentId = data.id;
        localStorage.setItem("sudoku.currentId", currentId);
        dirty=false;
        if(!silent) alert("Saved with id: "+data.id);
      }else{
        const msg = "Save failed: "+(data.error||"unknown");
        if(!silent) alert(msg); else console.warn(msg);
      }
    }catch(e){
      const msg = "Save error: "+e;
      if(!silent) alert(msg); else console.warn(msg);
    }
  }
  function markDirty(){
    dirty=true;
    scheduleAutosave();
  }
  function scheduleAutosave(){
    if(autosaveTimer) clearTimeout(autosaveTimer);
    autosaveTimer = setTimeout(()=>{ if(dirty) savePuzzle({silent:true}); }, AUTOSAVE_DEBOUNCE_MS);
  }
  // periodic safety autosave
  setInterval(()=>{ if(dirty) savePuzzle({silent:true}); }, AUTOSAVE_INTERVAL_MS);
  // before unload
  window.addEventListener("beforeunload", (e)=>{
    if(dirty){
      // try to save but don't block the unload
      navigator.sendBeacon && navigator.sendBeacon("/api/save", new Blob([JSON.stringify({
        id: currentId || "",
        name: (nameInput?.value||"").trim(),
        notes: (notesInput?.value||"").trim(),
        board:{board:getBoard(), fixed:getFixed()}
      })], {type:"application/json"}));
    }
  });

  // Manual save button now uses shared save
  btnSave?.addEventListener("click",()=>savePuzzle({silent:false}));

  btnLoad?.addEventListener("click",async()=>{
    try{
      const lr=await fetch("/api/list");
      const l=await lr.json();
      const items=(l.puzzles||[]);
      if(items.length===0){ alert("No saved puzzles."); return; }
      const lines=items.map(p=> p.name ? `${p.id} — ${p.name}` : p.id);
      const choice=prompt("Enter id to load:\n"+lines.join("\n"));
      if(!choice) return;
      const id=choice.split(" — ")[0].trim();
      const data=await api("/api/load",{id});
      if(data.puzzle){
        setBoard(data.puzzle.board.board, data.puzzle.board.fixed);
        nameInput.value=data.puzzle.name||"";
        notesInput.value=data.puzzle.notes||"";
        // difficulty handling is optional; keep current selector value if not provided
        currentId = data.puzzle.id || id;
        localStorage.setItem("sudoku.currentId", currentId);
        clearClass("conflict"); clearInlineOutlines();
        undoStack.length=0; redoStack.length=0;
        dirty=false;
      } else { alert("Load failed: "+(data.error||"unknown")); }
    }catch(e){ alert("Load error: "+e); }
  });

  autoCand?.addEventListener("change",()=>{ renderAll(); });

  // initial paint
  renderAll();
  highlight();
});