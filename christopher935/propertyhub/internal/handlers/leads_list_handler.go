package handlers
import ("github.com/gin-gonic/gin"; "gorm.io/gorm"; "chrisgross-ctrl-project/internal/security"; "net/http")
type LeadsListHandler struct {db *gorm.DB; encryptionManager *security.EncryptionManager}
func NewLeadsListHandler(db *gorm.DB, encMgr *security.EncryptionManager) *LeadsListHandler {return &LeadsListHandler{db: db, encryptionManager: encMgr}}
func (h *LeadsListHandler) GetAllLeads(c *gin.Context) {
	var result []map[string]interface{}
	h.db.Raw("SELECT id, name, email, phone, status, COALESCE(behavioral_score, 30) as behavioral_score, 'cold' as segment FROM leads LIMIT 50").Scan(&result)
	c.JSON(http.StatusOK, gin.H{"success": true, "leads": result, "pagination": gin.H{"page": 1, "limit": 50, "total": len(result)}, "stats": gin.H{"total": len(result), "hot": 0, "warm": 0, "cold": 1}})
}
