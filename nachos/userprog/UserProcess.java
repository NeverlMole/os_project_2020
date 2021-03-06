package nachos.userprog;

import nachos.machine.*;
import nachos.threads.*;
import nachos.userprog.*;

import java.io.EOFException;
import java.util.Iterator;
import java.util.LinkedList;

/**
 * Encapsulates the state of a user process that is not contained in its
 * user thread (or threads). This includes its address translation state, a
 * file table, and information about the program being executed.
 *
 * <p>
 * This class is extended by other classes to support additional functionality
 * (such as additional syscalls).
 *
 * @see	nachos.vm.VMProcess
 * @see	nachos.network.NetProcess
 */
public class UserProcess {
  /**
   * Allocate a new process.
   */
  public UserProcess() {
    for(int i=0; i<16; i++) {
      FileDescr[i] = null;
    }
  }

  /**
   * Allocate and return a new process of the correct class. The class name
   * is specified by the <tt>nachos.conf</tt> key
   * <tt>Kernel.processClassName</tt>.
   *
   * @return	a new process of the correct class.
   */
  public static UserProcess newUserProcess() {
    return (UserProcess)Lib.constructObject(Machine.getProcessClassName());
  }

  /**
   * Execute the specified program with the specified arguments. Attempts to
   * load the program, and then forks a thread to run it.
   *
   * @param	name	the name of the file containing the executable.
   * @param	args	the arguments to pass to the executable.
   * @return	<tt>true</tt> if the program was successfully executed.
   */
  public boolean execute(String name, String[] args) {
    if (!load(name, args))
      return false;

    thread = new UThread(this);

    // init processID.
    processID = nextProcessID;
    nextProcessID += 1;
    numRunningProcess += 1;
    if (processID == 0) {
      isRoot = true;
    }

    // init childList
    childList = new LinkedList<UserProcess>();

    Lib.debug(dbgProcess, "Exec " + name + " with the following arguments:");
    for (int i = 0; i < args.length; i++) {
      Lib.debug(dbgProcess, " - " + args[i]);
    }

    // init stdin, stdout.
    FileDescr[0] = UserKernel.console.openForReading();
    FileDescr[1] = UserKernel.console.openForWriting();

    thread.setName(name).fork();

    return true;
  }

  /**
   * Save the state of this process in preparation for a context switch.
   * Called by <tt>UThread.saveState()</tt>.
   */
  public void saveState() {}

  /**
   * Restore the state of this process after a context switch. Called by
   * <tt>UThread.restoreState()</tt>.
   */
  public void restoreState() { Machine.processor().setPageTable(pageTable); }

  /**
   * Read a null-terminated string from this process's virtual memory. Read
   * at most <tt>maxLength + 1</tt> bytes from the specified address, search
   * for the null terminator, and convert it to a <tt>java.lang.String</tt>,
   * without including the null terminator. If no null terminator is found,
   * returns <tt>null</tt>.
   *
   * @param	vaddr	the starting virtual address of the null-terminated
   *			string.
   * @param	maxLength	the maximum number of characters in the string,
   *				not including the null terminator.
   * @return	the string read, or <tt>null</tt> if no null terminator was
   *		found.
   */
  public String readVirtualMemoryString(int vaddr, int maxLength) {
    Lib.assertTrue(maxLength >= 0);

    byte[] bytes = new byte[maxLength + 1];

    int bytesRead = readVirtualMemory(vaddr, bytes);

    for (int length = 0; length < bytesRead; length++) {
      if (bytes[length] == 0)
        return new String(bytes, 0, length);
    }

    return null;
  }

  /**
   * Read an memory address from a virtual memory address. Return -1 if fails.
   */
  public int readVirtualMemoryAddr(int vaddr) {
    byte[] tmp=new byte[AddrMemoryLength];

    if (readVirtualMemory(vaddr, tmp) != AddrMemoryLength) {
      return -1;
    }

    return Lib.bytesToInt(tmp, 0);
  }

  /**
   * Transfer data from this process's virtual memory to all of the specified
   * array. Same as <tt>readVirtualMemory(vaddr, data, 0, data.length)</tt>.
   *
   * @param	vaddr	the first byte of virtual memory to read.
   * @param	data	the array where the data will be stored.
   * @return	the number of bytes successfully transferred.
   */
  public int readVirtualMemory(int vaddr, byte[] data) {
    return readVirtualMemory(vaddr, data, 0, data.length);
  }

  /**
   * Transfer data from this process's virtual memory to the specified array.
   * This method handles address translation details. This method must
   * <i>not</i> destroy the current process if an error occurs, but instead
   * should return the number of bytes successfully copied (or zero if no
   * data could be copied).
   *
   * @param	vaddr	the first byte of virtual memory to read.
   * @param	data	the array where the data will be stored.
   * @param	offset	the first byte to write in the array.
   * @param	length	the number of bytes to transfer from virtual memory to
   *			the array.
   * @return	the number of bytes successfully transferred.
   */
  public int readVirtualMemory(int vaddr, byte[] data, int offset, int length) {
    Lib.assertTrue(offset >= 0 && length >= 0 &&
                   offset + length <= data.length);

    if (vaddr < 0) {
      return 0;
    }

    int vpn = Processor.pageFromAddress(vaddr);
    int offsetv = Processor.offsetFromAddress(vaddr);

    byte[] memory = Machine.processor().getMemory();

    int amount = 0;

    while (vpn < numPages) {
      int copyLength = Math.min(pageSize - offsetv, length);
      int copyAddr = Processor.makeAddress(pageTable[vpn].ppn, offsetv);

      System.arraycopy(memory, copyAddr, data, offset + amount, copyLength);

      amount += copyLength;
      vpn += 1;
      offsetv = 0;
      length -= copyLength;

      if (length == 0) {
        break;
      }
    }

    return amount;
  }

  /**
   * Transfer all data from the specified array to this process's virtual
   * memory.
   * Same as <tt>writeVirtualMemory(vaddr, data, 0, data.length)</tt>.
   *
   * @param	vaddr	the first byte of virtual memory to write.
   * @param	data	the array containing the data to transfer.
   * @return	the number of bytes successfully transferred.
   */
  public int writeVirtualMemory(int vaddr, byte[] data) {
    return writeVirtualMemory(vaddr, data, 0, data.length);
  }

  /**
   * Transfer data from the specified array to this process's virtual memory.
   * This method handles address translation details. This method must
   * <i>not</i> destroy the current process if an error occurs, but instead
   * should return the number of bytes successfully copied (or zero if no
   * data could be copied).
   *
   * @param	vaddr	the first byte of virtual memory to write.
   * @param	data	the array containing the data to transfer.
   * @param	offset	the first byte to transfer from the array.
   * @param	length	the number of bytes to transfer from the array to
   *			virtual memory.
   * @return	the number of bytes successfully transferred.
   */
  public int writeVirtualMemory(int vaddr, byte[] data, int offset,
                                int length) {
    Lib.assertTrue(offset >= 0 && length >= 0 &&
                   offset + length <= data.length);

    if (vaddr < 0) {
      return 0;
    }

    int vpn = Processor.pageFromAddress(vaddr);
    int offsetv = Processor.offsetFromAddress(vaddr);

    byte[] memory = Machine.processor().getMemory();

    int amount = 0;

    while (vpn < numPages) {
      if (pageTable[vpn].readOnly) {
        break;
      }

      int copyLength = Math.min(pageSize - offsetv, length);
      int copyAddr = Processor.makeAddress(pageTable[vpn].ppn, offsetv);

      System.arraycopy(data, offset + amount, memory, copyAddr, copyLength);

      amount += copyLength;
      vpn += 1;
      offsetv = 0;
      length -= copyLength;

      if (length == 0) {
        break;
      }
    }

    return amount;
  }

  /**
   * Load the executable with the specified name into this process, and
   * prepare to pass it the specified arguments. Opens the executable, reads
   * its header information, and copies sections and arguments into this
   * process's virtual memory.
   *
   * @param	name	the name of the file containing the executable.
   * @param	args	the arguments to pass to the executable.
   * @return	<tt>true</tt> if the executable was successfully loaded.
   */
  private boolean load(String name, String[] args) {
    Lib.debug(dbgProcess, "UserProcess.load(\"" + name + "\")");

    OpenFile executable = ThreadedKernel.fileSystem.open(name, false);
    if (executable == null) {
      Lib.debug(dbgProcess, "\topen failed");
      return false;
    }

    try {
      coff = new Coff(executable);
    } catch (EOFException e) {
      executable.close();
      Lib.debug(dbgProcess, "\tcoff load failed");
      return false;
    }

    // make sure the sections are contiguous and start at page 0
    numPages = 0;
    for (int s = 0; s < coff.getNumSections(); s++) {
      CoffSection section = coff.getSection(s);
      if (section.getFirstVPN() != numPages) {
        coff.close();
        Lib.debug(dbgProcess, "\tfragmented executable");
        return false;
      }
      numPages += section.getLength();
    }

    // make sure the argv array will fit in one page
    byte[][] argv = new byte[args.length][];
    int argsSize = 0;
    for (int i = 0; i < args.length; i++) {
      argv[i] = args[i].getBytes();
      // 4 bytes for argv[] pointer; then string plus one for null byte
      argsSize += 4 + argv[i].length + 1;
    }
    if (argsSize > pageSize) {
      coff.close();
      Lib.debug(dbgProcess, "\targuments too long");
      return false;
    }

    // program counter initially points at the program entry point
    initialPC = coff.getEntryPoint();

    // next comes the stack; stack pointer initially points to top of it
    numPages += stackPages;
    initialSP = numPages * pageSize;

    // and finally reserve 1 page for arguments
    numPages++;

    if (!loadSections())
      return false;

    // store arguments in last page
    int entryOffset = (numPages - 1) * pageSize;
    int stringOffset = entryOffset + args.length * 4;

    this.argc = args.length;
    this.argv = entryOffset;

    for (int i = 0; i < argv.length; i++) {
      byte[] stringOffsetBytes = Lib.bytesFromInt(stringOffset);
      Lib.assertTrue(writeVirtualMemory(entryOffset, stringOffsetBytes) == 4);
      entryOffset += 4;
      Lib.assertTrue(writeVirtualMemory(stringOffset, argv[i]) ==
                     argv[i].length);
      stringOffset += argv[i].length;
      Lib.assertTrue(writeVirtualMemory(stringOffset, new byte[] {0}) == 1);
      stringOffset += 1;
    }

    return true;
  }

  /**
   * Allocates memory for this process, and loads the COFF sections into
   * memory. If this returns successfully, the process will definitely be
   * run (this is the last step in process initialization that can fail).
   *
   * @return	<tt>true</tt> if the sections were successfully loaded.
   */
  protected boolean loadSections() {
    pageTable = new TranslationEntry[numPages];
    for (int i = 0; i < numPages; i++) {
      int ppn = UserKernel.allocatePage();

      if (ppn == -1) {
        Lib.debug(dbgProcess, "\tinsufficient physical memory");
        i -= 1;
        while (i >= 0) {
          UserKernel.freePage(pageTable[i].ppn);
        }
        return false;
      }

      pageTable[i] = new TranslationEntry(i, ppn, true, false, false, false);
    }

    // load sections
    for (int s = 0; s < coff.getNumSections(); s++) {
      CoffSection section = coff.getSection(s);

      Lib.debug(dbgProcess, "\tinitializing " + section.getName() +
                              " section (" + section.getLength() + " pages)");

      for (int i = 0; i < section.getLength(); i++) {
        int vpn = section.getFirstVPN() + i;
        pageTable[vpn].readOnly = section.isReadOnly();

        // for now, just assume virtual addresses=physical addresses
        section.loadPage(i, pageTable[vpn].ppn);
      }
    }

    return true;
  }

  /**
   * Release any resources allocated by <tt>loadSections()</tt>.
   */
  protected void unloadSections() {
    for (int i = 0; i < numPages; i++) {
      UserKernel.freePage(pageTable[i].ppn);
    }
  }

  private void finish() {
    coff.close();
    for (int i = 0; i < 16; i++) {
      if (FileDescr[i] != null) {
        FileDescr[i].close();
        FileDescr[i] = null;
      }
    }
    unloadSections();

    if (numRunningProcess == 1) {
      Lib.debug(dbgProcess,
                "Kernel terminated because of the last process finished.");
      UserKernel.kernel.terminate();
    } else {
      numRunningProcess -= 1;
      thread.finish();
    }

    Lib.assertNotReached("Process did not exit after finished.");
  }

  /**
   * Initialize the processor's registers in preparation for running the
   * program loaded into this process. Set the PC register to point at the
   * start function, set the stack pointer register to point at the top of
   * the stack, set the A0 and A1 registers to argc and argv, respectively,
   * and initialize all other registers to 0.
   */
  public void initRegisters() {
    Processor processor = Machine.processor();

    // by default, everything's 0
    for (int i = 0; i < processor.numUserRegisters; i++)
      processor.writeRegister(i, 0);

    // initialize PC and SP according
    processor.writeRegister(Processor.regPC, initialPC);
    processor.writeRegister(Processor.regSP, initialSP);

    // initialize the first two argument registers to argc and argv
    processor.writeRegister(Processor.regA0, argc);
    processor.writeRegister(Processor.regA1, argv);
  }


/* New codes for halt, create, open, read, write, close, unlink between these lines*/

  private int handleHalt() {
    if (!isRoot) {
      return 0;
    }

    Machine.halt();

    Lib.assertNotReached("Machine.halt() did not halt machine!");
    return 0;
  }

  private int handleCreate(int a0) {
    return handleCO(a0, true);
  }
  private int handleOpen(int a0) {
    return handleCO(a0, false);
  }

  private int handleCO(int a0, boolean crea){
    String Filename = readVirtualMemoryString(a0, 256);
    Lib.debug(dbgProcess, "UserProcess.CO(\"" + Filename + "\")");
    if(Filename == null) {
      Lib.debug(dbgProcess, "\tfilename empty");
      return -1;
    }
    int loc = AlloFileDescr();
    if(loc == -1) return -1;
    OpenFile newFile = ThreadedKernel.fileSystem.open(Filename, crea);
    if (newFile == null) {
      Lib.debug(dbgProcess, "\tcreate failed");
      return -1;
    }
    FileDescr[loc] = newFile;
    return loc;
  }

  private int handleRead(int a0, int a1, int a2) {
    Lib.debug(dbgProcess, "UserProcess.Read()");

    if (!checkDescr(a0)) return -1;
    byte[] buff = new byte[a2];
    int tryread = FileDescr[a0].read(buff, 0, a2);

    Lib.debug(dbgProcess, "Read bits " + tryread + ".");

    if (tryread == -1) {
      Lib.debug(dbgProcess, "\tread failed");
      return -1;
    }
    int trywrite = writeVirtualMemory(a1, buff, 0, tryread);

    if (trywrite < tryread) {
      return -1;
    }
    return trywrite;
  }

  private int handleWrite(int a0, int a1, int a2) {
    Lib.debug(dbgProcess, "UserProcess.Write()");
    if (!checkDescr(a0)) return -1;
    byte[] buff = new byte[a2];

    int tryread = readVirtualMemory(a1, buff);
    if(tryread != a2){ 
      Lib.debug(dbgProcess, "\tload failed");
      return -1;
    }
    int trywrite = FileDescr[a0].write(buff, 0, a2);
    if(trywrite != a2){ 
      Lib.debug(dbgProcess, "\twrite incomplete");
      return -1;
    }
    return trywrite;
  }

  private int handleClose(int a0) {
    Lib.debug(dbgProcess, "UserProcess.Close()");
    if (!checkDescr(a0)) return -1;
    FileDescr[a0].close();
    FileDescr[a0] = null;
    return 0;
  }

  private int handleUnlink(int a0) {
    String Filename = readVirtualMemoryString(a0, 255);
    Lib.debug(dbgProcess, "UserProcess.Unlink(\"" + Filename + "\")");
    if (Filename == null) return 0;
    boolean res = ThreadedKernel.fileSystem.remove(Filename);
    if(res) return 0; else return -1;
  }

  private boolean checkDescr(int a0){
    if (a0*(15-a0)<0){
      Lib.debug(dbgProcess, "\tinvalid descriptor");
      return false;
    }
    if (FileDescr[a0] == null){
      Lib.debug(dbgProcess, "\tnull descriptor");
      return false;
    }
    return true;
  }


  private int AlloFileDescr(){
    for(int i=0; i<16; i++)
      if(FileDescr[i]==null) return i;
    Lib.debug(dbgProcess, "\ttoo many files");
    return -1;
  }


  /**
   * Handle the exec() system call.
   */
  private int handleExec(int fileAddr, int argc, int argvAddr) {
    if (argc < 0) {
      return -1;
    }

    String file = readVirtualMemoryString(fileAddr, MaxArgLength);
    String[] args = new String[argc];

    for(int i = 0; i < argc ;i++) {
      int argaddr = readVirtualMemoryAddr(argvAddr + AddrMemoryLength * i);
      args[i] = readVirtualMemoryString(argaddr, MaxArgLength);
    }

    UserProcess process = UserProcess.newUserProcess();

    if (!process.execute(file, args)) {
      return -1;
    }

    childList.add(process);

    return process.processID;
  }

  /**
   * Handle the exit() system call.
   */
  private int handleExit(int status) {
    normallyExit = true;
    returnStatus = status;
    finish();
    return 0;
  }

  /**
   * Handle the join() system call.
   */
  private int handleJoin(int pid, int statusAddr) {
    UserProcess cprocess = null;

    for (Iterator it = childList.iterator(); it.hasNext();) {
      UserProcess process = (UserProcess) it.next();
      if (process.processID == pid) {
        cprocess = process;
      }
    }

    if (cprocess == null) {
      return -1;
    }

    cprocess.thread.join();

    if (!cprocess.normallyExit) {
      return 0;
    }

    writeVirtualMemory(statusAddr, Lib.bytesFromInt(cprocess.returnStatus));
    return 1;
  }

  private static final int syscallHalt = 0, syscallExit = 1, syscallExec = 2,
                           syscallJoin = 3, syscallCreate = 4, syscallOpen = 5,
                           syscallRead = 6, syscallWrite = 7, syscallClose = 8,
                           syscallUnlink = 9;

  /**
   * Handle a syscall exception. Called by <tt>handleException()</tt>. The
   * <i>syscall</i> argument identifies which syscall the user executed:
   *
   * <table>
   * <tr><td>syscall#</td><td>syscall prototype</td></tr>
   * <tr><td>0</td><td><tt>void halt();</tt></td></tr>
   * <tr><td>1</td><td><tt>void exit(int status);</tt></td></tr>
   * <tr><td>2</td><td><tt>int  exec(char *name, int argc, char **argv);
   * 								</tt></td></tr>
   * <tr><td>3</td><td><tt>int  join(int pid, int *status);</tt></td></tr>
   * <tr><td>4</td><td><tt>int  creat(char *name);</tt></td></tr>
   * <tr><td>5</td><td><tt>int  open(char *name);</tt></td></tr>
   * <tr><td>6</td><td><tt>int  read(int fd, char *buffer, int size);
   *								</tt></td></tr>
   * <tr><td>7</td><td><tt>int  write(int fd, char *buffer, int size);
   *								</tt></td></tr>
   * <tr><td>8</td><td><tt>int  close(int fd);</tt></td></tr>
   * <tr><td>9</td><td><tt>int  unlink(char *name);</tt></td></tr>
   * </table>
   *
   * @param	syscall	the syscall number.
   * @param	a0	the first syscall argument.
   * @param	a1	the second syscall argument.
   * @param	a2	the third syscall argument.
   * @param	a3	the fourth syscall argument.
   * @return	the value to be returned to the user.
   */
  public int handleSyscall(int syscall, int a0, int a1, int a2, int a3) {
    switch (syscall) {
    case syscallHalt:
      return handleHalt();
    case syscallCreate:
      return handleCreate(a0);
    case syscallOpen:
      return handleOpen(a0);
    case syscallWrite:
      return handleWrite(a0,a1,a2);
    case syscallRead:
      return handleRead(a0,a1,a2);
    case syscallClose:
      return handleClose(a0);
    case syscallUnlink:
      return handleUnlink(a0);
    case syscallExit:
      return handleExit(a0);
    case syscallExec:
      return handleExec(a0, a1, a2);
    case syscallJoin:
      return handleJoin(a0, a1);

    default:
      Lib.debug(dbgProcess, "Unknown syscall " + syscall);
      finish();
      Lib.assertNotReached("Unknown system call!");
    }
    return 0;
  }

  /**
   * Handle a user exception. Called by
   * <tt>UserKernel.exceptionHandler()</tt>. The
   * <i>cause</i> argument identifies which exception occurred; see the
   * <tt>Processor.exceptionZZZ</tt> constants.
   *
   * @param	cause	the user exception that occurred.
   */
  public void handleException(int cause) {
    Processor processor = Machine.processor();

    switch (cause) {
    case Processor.exceptionSyscall:
      int result = handleSyscall(processor.readRegister(Processor.regV0),
                                 processor.readRegister(Processor.regA0),
                                 processor.readRegister(Processor.regA1),
                                 processor.readRegister(Processor.regA2),
                                 processor.readRegister(Processor.regA3));
      processor.writeRegister(Processor.regV0, result);
      processor.advancePC();
      break;

    default:
      Lib.debug(dbgProcess,
                "Unexpected exception: " + Processor.exceptionNames[cause]);
      finish();
      Lib.assertNotReached("Unexpected exception");
    }
  }

  /** Selftest. */

  public static void memoryTest() {
    System.out.println("Enter memoryTest for UserPrcross.");
    UserProcess process = UserProcess.newUserProcess();

    process.numPages = 5;

    process.pageTable = new TranslationEntry[5];

    for (int i = 0; i < process.numPages; i++) {
      int ppn = UserKernel.allocatePage();

      Lib.assertTrue(ppn >= 0);

      process.pageTable[i] =
        new TranslationEntry(i, ppn, true, false, false, false);
    }

    process.pageTable[2].readOnly = true;

    // Test normal read and write.
    int voffset = 10;
    int vpn = 0;
    String mark = "Here Here!";
    byte[] tmp = mark.getBytes();

    int numbw = process.writeVirtualMemory(Processor.makeAddress(vpn, voffset),
                                           tmp);

    System.out.println("Write to (" + vpn + "," + voffset + ") string "
                       + mark + " with " + numbw + ".");

    byte[] tmp2 = new byte[5000];
    int numbr = process.readVirtualMemory(Processor.makeAddress(vpn, voffset),
        tmp2, 0, 500);

    System.out.println("Read from (" + vpn + "," + voffset + ") " + numbr
                       + " bits.");

    // Test read and write to readonly pages.
    voffset = pageSize - 5;
    vpn = 1;

    numbw = process.writeVirtualMemory(Processor.makeAddress(vpn, voffset),
                                           tmp);

    System.out.println("Write to (" + vpn + "," + voffset + ") string "
                       + mark + " with " + numbw + ".");

    numbr = process.readVirtualMemory(Processor.makeAddress(vpn, voffset),
        tmp2, 0, 500);

    System.out.println("Read from (" + vpn + "," + voffset + ") " + numbr
                       + " bits .");

    // Test read and write when out of virtul memory.
    voffset = pageSize - 5;
    vpn = 4;

    numbw = process.writeVirtualMemory(Processor.makeAddress(vpn, voffset),
                                           tmp);

    System.out.println("Write to (" + vpn + "," + voffset + ") string "
                       + mark + " with " + numbw + ".");

    numbr = process.readVirtualMemory(Processor.makeAddress(vpn, voffset),
        tmp2, 0, 500);

    System.out.println("Read from (" + vpn + "," + voffset + ") " + numbr
                       + " bits.");

    mark = "";

    for (int i = 0; i < 500; i++) {
      mark = mark + i;
    }
    tmp = mark.getBytes();

    // Test for long read and write.
    voffset = 0;
    vpn = 0;

    numbw = process.writeVirtualMemory(Processor.makeAddress(vpn, voffset),
                                           tmp);

    System.out.println("Write to (" + vpn + "," + voffset + ") string "
                       + mark + " with " + numbw + ".");

    vpn += 1;
    String sread = process.readVirtualMemoryString(
        Processor.makeAddress(vpn, voffset), 2000);

    System.out.println("Read from (" + vpn + "," + voffset + ") string "
                       + sread);

    for (int i = 0; i < process.numPages; i++) {
      UserKernel.freePage(process.pageTable[i].ppn);
    }
  }

  public static void selfTest() {
    memoryTest();
  }

  /** The program being run by this process. */
  protected Coff coff;

  /** This process's page table. */
  protected TranslationEntry[] pageTable;
  /** The number of contiguous pages occupied by the program. */
  protected int numPages;

  /** The number of pages in the program's stack. */
  protected final int stackPages = 8;

  private int initialPC, initialSP;
  private int argc, argv;


  private static final int pageSize = Processor.pageSize;
  private static final char dbgProcess = 'a';

  /** variables for dealing root process and process id. */
  private boolean isRoot = false;
  private int processID;
  private static int nextProcessID = 0;
  private static int numRunningProcess = 0;

  /** the thread associated with the process. */
  private UThread thread;

  /** child processes of a the process. */
  private LinkedList<UserProcess> childList;

  /** return states. */
  private int returnStatus;
  private boolean normallyExit = false;

  /** file descriptors. */
  private OpenFile[] FileDescr = new OpenFile[16];

  private static final int AddrMemoryLength = 4;
  private static final int MaxArgLength = 256;
}
