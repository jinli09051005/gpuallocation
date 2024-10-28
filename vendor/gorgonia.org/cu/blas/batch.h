#include <stdio.h>
#include <cuda.h>
#include <cublas_v2.h>

typedef enum {
	fn_undefined,

	fn_cublasSnrm2, 
	fn_cublasDnrm2, 
	fn_cublasScnrm2, 
	fn_cublasDznrm2, 
	fn_cublasSdot, 
	fn_cublasDdot, 
	fn_cublasSscal, 
	fn_cublasDscal, 
	fn_cublasCscal, 
	fn_cublasCsscal, 
	fn_cublasZscal, 
	fn_cublasZdscal, 
	fn_cublasSaxpy, 
	fn_cublasDaxpy, 
	fn_cublasCaxpy, 
	fn_cublasZaxpy, 
	fn_cublasScopy, 
	fn_cublasDcopy, 
	fn_cublasCcopy, 
	fn_cublasZcopy, 
	fn_cublasSswap, 
	fn_cublasDswap, 
	fn_cublasCswap, 
	fn_cublasZswap, 
	fn_cublasIsamax, 
	fn_cublasIdamax, 
	fn_cublasIcamax, 
	fn_cublasIzamax, 
	fn_cublasIsamin, 
	fn_cublasIdamin, 
	fn_cublasIcamin, 
	fn_cublasIzamin, 
	fn_cublasSasum, 
	fn_cublasDasum, 
	fn_cublasScasum, 
	fn_cublasDzasum, 
	fn_cublasSrot, 
	fn_cublasDrot, 
	fn_cublasCrot, 
	fn_cublasZrot, 
	fn_cublasSgemv, 
	fn_cublasDgemv, 
	fn_cublasCgemv, 
	fn_cublasZgemv, 
	fn_cublasSgbmv, 
	fn_cublasDgbmv, 
	fn_cublasCgbmv, 
	fn_cublasZgbmv, 
	fn_cublasStrmv, 
	fn_cublasDtrmv, 
	fn_cublasCtrmv, 
	fn_cublasZtrmv, 
	fn_cublasStbmv, 
	fn_cublasDtbmv, 
	fn_cublasCtbmv, 
	fn_cublasZtbmv, 
	fn_cublasStpmv, 
	fn_cublasDtpmv, 
	fn_cublasCtpmv, 
	fn_cublasZtpmv, 
	fn_cublasStrsv, 
	fn_cublasDtrsv, 
	fn_cublasCtrsv, 
	fn_cublasZtrsv, 
	fn_cublasStpsv, 
	fn_cublasDtpsv, 
	fn_cublasCtpsv, 
	fn_cublasZtpsv, 
	fn_cublasStbsv, 
	fn_cublasDtbsv, 
	fn_cublasCtbsv, 
	fn_cublasZtbsv, 
	fn_cublasSsymv, 
	fn_cublasDsymv, 
	fn_cublasCsymv, 
	fn_cublasZsymv, 
	fn_cublasChemv, 
	fn_cublasZhemv, 
	fn_cublasSsbmv, 
	fn_cublasDsbmv, 
	fn_cublasChbmv, 
	fn_cublasZhbmv, 
	fn_cublasSspmv, 
	fn_cublasDspmv, 
	fn_cublasChpmv, 
	fn_cublasZhpmv, 
	fn_cublasSger, 
	fn_cublasDger, 
	fn_cublasCgeru, 
	fn_cublasCgerc, 
	fn_cublasZgeru, 
	fn_cublasZgerc, 
	fn_cublasSsyr, 
	fn_cublasDsyr, 
	fn_cublasCsyr, 
	fn_cublasZsyr, 
	fn_cublasCher, 
	fn_cublasZher, 
	fn_cublasSspr, 
	fn_cublasDspr, 
	fn_cublasChpr, 
	fn_cublasZhpr, 
	fn_cublasSsyr2, 
	fn_cublasDsyr2, 
	fn_cublasCsyr2, 
	fn_cublasZsyr2, 
	fn_cublasCher2, 
	fn_cublasZher2, 
	fn_cublasSspr2, 
	fn_cublasDspr2, 
	fn_cublasChpr2, 
	fn_cublasZhpr2, 
	fn_cublasSgemm, 
	fn_cublasDgemm, 
	fn_cublasCgemm, 
	fn_cublasCgemm3m, 
	fn_cublasZgemm, 
	fn_cublasZgemm3m, 
	fn_cublasSsyrk, 
	fn_cublasDsyrk, 
	fn_cublasCsyrk, 
	fn_cublasZsyrk, 
	fn_cublasCherk, 
	fn_cublasZherk, 
	fn_cublasSsyr2k, 
	fn_cublasDsyr2k, 
	fn_cublasCsyr2k, 
	fn_cublasZsyr2k, 
	fn_cublasCher2k, 
	fn_cublasZher2k, 
	fn_cublasSsyrkx, 
	fn_cublasDsyrkx, 
	fn_cublasCsyrkx, 
	fn_cublasZsyrkx, 
	fn_cublasCherkx, 
	fn_cublasZherkx, 
	fn_cublasSsymm, 
	fn_cublasDsymm, 
	fn_cublasCsymm, 
	fn_cublasZsymm, 
	fn_cublasChemm, 
	fn_cublasZhemm, 
	fn_cublasStrsm, 
	fn_cublasDtrsm, 
	fn_cublasCtrsm, 
	fn_cublasZtrsm, 
	fn_cublasSgeam, 
	fn_cublasDgeam, 
	fn_cublasCgeam, 
	fn_cublasZgeam, 
	fn_cublasSdgmm, 
	fn_cublasDdgmm, 
	fn_cublasCdgmm, 
	fn_cublasZdgmm, 
	fn_cublasStpttr, 
	fn_cublasDtpttr, 
	fn_cublasCtpttr, 
	fn_cublasZtpttr, 
	fn_cublasStrttp, 
	fn_cublasDtrttp, 
	fn_cublasCtrttp, 
	fn_cublasZtrttp, 
	
} cublasFn;
